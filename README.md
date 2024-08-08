Kubectl plugin for parallel applying based on common label. Default label is parallel/group, but it can be changed with `--label` flag.

Read [using plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/#using-a-plugin) on how to use it with kubectl.
## Usage
```sh
cat test.yaml | kubectl parallel apply -f - -l parallel/group
```
