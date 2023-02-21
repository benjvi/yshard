package main

/*type inputIter interface {
	gojq.Iter
	io.Closer
	Name() string
}

func runYAML() {
	iter := createInputIter(args)
	defer iter.Close()
	code, err := gojq.Compile(query,
		gojq.WithModuleLoader(gojq.NewModuleLoader(modulePaths)),
		gojq.WithEnvironLoader(os.Environ),
		gojq.WithVariables(cli.argnames),
		gojq.WithFunction("debug", 0, 0, cli.funcDebug),
		gojq.WithFunction("stderr", 0, 0, cli.funcStderr),
		gojq.WithFunction("input_filename", 0, 0,
			func(iter inputIter) func(any, []any) any {
				return func(any, []any) any {
					if fname := iter.Name(); fname != "" && (len(args) > 0 || !opts.InputNull) {
						return fname
					}
					return nil
				}
			}(iter),
		),
		gojq.WithInputIter(iter),
	)
	if err != nil {
		if err, ok := err.(interface {
			QueryParseError() (string, string, error)
		}); ok {
			name, query, err := err.QueryParseError()
			return &queryParseError{name, query, err}
		}
		if err, ok := err.(interface {
			JSONParseError() (string, string, error)
		}); ok {
			fname, contents, err := err.JSONParseError()
			return &compileError{&jsonParseError{fname, contents, 0, err}}
		}
		return &compileError{err}
	}
	if opts.InputNull {
		iter = newNullInputIter()
	}
	return process(iter, code)
}

func (cli *cli) process(iter inputIter, code *gojq.Code) error {
	var err error
	for {
		v, ok := iter.Next()
		if !ok {
			return err
		}
		if er, ok := v.(error); ok {
			cli.printError(er)
			err = &emptyError{er}
			continue
		}
		if er := cli.printValues(code.Run(v, cli.argvalues...)); er != nil {
			cli.printError(er)
			err = &emptyError{er}
		}
	}
}


func createInputIter(args []string) (iter inputIter) {
	var newIter func(io.Reader, string) inputIter
	newIter = newYAMLInputIter
	return newFilesInputIter(newIter, args, cli.inStream)
}

func newFilesInputIter(
	newIter func(io.Reader, string) inputIter, fnames []string, stdin io.Reader,
) inputIter {
	return &filesInputIter{newIter: newIter, fnames: fnames, stdin: stdin}
}

func newInputReader(r io.Reader) *inputReader {
	if r, ok := r.(*os.File); ok {
		if _, err := r.Seek(0, io.SeekCurrent); err == nil {
			return &inputReader{r, r, nil}
		}
	}
	var buf bytes.Buffer // do not use strings.Builder because we need to Reset
	return &inputReader{io.TeeReader(r, &buf), nil, &buf}
}

func newYAMLInputIter(r io.Reader, fname string) inputIter {
	ir := newInputReader(r)
	dec := yaml.NewDecoder(ir)
	return &yamlInputIter{dec: dec, ir: ir, fname: fname}
}

type inputReader struct {
	io.Reader
	file *os.File
	buf  *bytes.Buffer
}

type yamlInputIter struct {
	dec   *yaml.Decoder
	ir    *inputReader
	fname string
	err   error
}

func (i *yamlInputIter) Next() (any, bool) {
	if i.err != nil {
		return nil, false
	}
	var v any
	if err := i.dec.Decode(&v); err != nil {
		if err == io.EOF {
			i.err = err
			return nil, false
		}
		i.err = &yamlParseError{i.fname, i.ir.getContents(nil, nil), err}
		return i.err, true
	}
	return normalizeYAML(v), true
}

func (i *yamlInputIter) Close() error {
	i.err = io.EOF
	return nil
}

func (i *yamlInputIter) Name() string {
	return i.fname
}

// Workaround for https://github.com/go-yaml/yaml/issues/139
func normalizeYAML(v any) any {
	switch v := v.(type) {
	case map[any]any:
		w := make(map[string]any, len(v))
		for k, v := range v {
			w[fmt.Sprint(k)] = normalizeYAML(v)
		}
		return w

	case map[string]any:
		w := make(map[string]any, len(v))
		for k, v := range v {
			w[k] = normalizeYAML(v)
		}
		return w

	case []any:
		for i, w := range v {
			v[i] = normalizeYAML(w)
		}
		return v

	// go-yaml unmarshals timestamp string to time.Time but gojq cannot handle it.
	// It is impossible to keep the original timestamp strings.
	case time.Time:
		return v.Format(time.RFC3339)

	default:
		return v
	}
}

type yamlParseError struct {
	fname, contents string
	err             error
}

func (err *yamlParseError) Error() string {
	var line int
	msg := strings.TrimPrefix(
		strings.TrimPrefix(err.err.Error(), "yaml: "),
		"unmarshal errors:\n  ")
	if fmt.Sscanf(msg, "line %d: ", &line); line == 0 {
		return "invalid yaml: " + err.fname
	}
	msg = msg[strings.Index(msg, ": ")+2:]
	if i := strings.IndexByte(msg, '\n'); i >= 0 {
		msg = msg[:i]
	}
	linestr := getLineByLine(err.contents, line)
	return fmt.Sprintf("invalid yaml: %s:%d\n%s  %s",
		err.fname, line, formatLineInfo(linestr, line, 0), msg)
}

func (ir *inputReader) getContents(offset *int64, line *int) string {
	if buf := ir.buf; buf != nil {
		return buf.String()
	}
	if current, err := ir.file.Seek(0, io.SeekCurrent); err == nil {
		defer func() { ir.file.Seek(current, io.SeekStart) }()
	}
	ir.file.Seek(0, io.SeekStart)
	const bufSize = 16 * 1024
	var buf bytes.Buffer // do not use strings.Builder because we need to Reset
	if offset != nil && *offset > bufSize {
		buf.Grow(bufSize)
		for *offset > bufSize {
			n, err := io.Copy(&buf, io.LimitReader(ir.file, bufSize))
			*offset -= int64(n)
			*line += bytes.Count(buf.Bytes(), []byte{'\n'})
			buf.Reset()
			if err != nil || n == 0 {
				break
			}
		}
	}
	var r io.Reader
	if offset == nil {
		r = ir.file
	} else {
		r = io.LimitReader(ir.file, bufSize*2)
	}
	io.Copy(&buf, r)
	return buf.String()
}
*/
