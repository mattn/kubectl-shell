# kubectl-shell

Open shell for the pod

## Usage

```
$ kubectl shell $(kubectl get pods --field-selector=status.phase=Running -o name | head -n 1)
```

## Installation

```
$ krew install shell
```

## License

MIT

## Author

Yasuhiro Matsumoto (a.k.a. mattn)
