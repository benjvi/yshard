# yshard

*Huge yaml files are impossible to understand, so let's split them up!*

yshard is a CLI that takes a single YAML file as input and splits it into separate files, doing a GROUP BY on the user-provided JSON path. It is particularly useful in cases where large complex packages are distributed as a single YAML file, containing multiple YAML documents - as is often the case for Kubernetes manifests. yshard is also useful when you render your own templates into a single output file. 


# Getting Started

Grab the binary for your architecture from [the releases page](https://github.com/benjvi/yshard/releases) 

Get a YAML file you want to split up, for example:

`wget https://raw.githubusercontent.com/argoproj/argo-workflows/stable/manifests/install.yaml`

Now use yshard to split the YAML, in this case we split by the `kind` field of each document, putting documents with the same `kind` in the same file in `output-directory`:

`cat install | yshard -g ".kind" -o output-directory`

Then in the output directory you see: 

```
$ ls output-directory
ClusterRole.yml              ConfigMap.yml                Deployment.yml               RoleBinding.yml              ServiceAccount.yml
ClusterRoleBinding.yml       CustomResourceDefinition.yml Role.yml                     Service.yml                  __ungrouped__.yml
```

