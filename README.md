Kubectl plugin for parallel applying based on common label. Default label is parallel/group, but it can be changed with `--label` flag.

## Usage
```sh
cat test.yaml | ./kubectl-parallel apply -f - -l parallel/group
```
