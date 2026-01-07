# kubectl-guard

Kubernetes context protection kubectl plugin.

Prevents accidental destructive operations on important contexts like production.

## Installation

### Homebrew

```bash
brew install sivchari/tap/kubectl-guard
```

### Krew

```bash
kubectl krew install guard
```

### Go install

```bash
go install github.com/sivchari/kubectl-guard/cmd/kubectl-guard@latest
```

### Build from source

```bash
git clone https://github.com/sivchari/kubectl-guard.git
cd kubectl-guard
make build
sudo mv kubectl-guard /usr/local/bin/
```

## Usage

### Guard a context

```bash
# Guard entire context
kubectl guard guard prod-cluster

# Guard specific namespaces only
kubectl guard guard prod-cluster --namespace=production,critical
```

### Unguard a context

```bash
kubectl guard unguard prod-cluster
```

### Check guard status

```bash
kubectl guard list
```

Output:
```
guarded contexts:
 * prod-cluster (all namespaces)
   staging-cluster (namespaces: production)

current: prod-cluster (guarded)
```

### Execute kubectl with protection check

```bash
# Blocked on guarded context
kubectl guard exec -- delete pod nginx

# Force execution with --force
kubectl guard exec -- delete pod nginx --force
```

## Configuration

Config is stored at `~/.kube/guard.yaml`.

```yaml
guardedContexts:
  - name: prod-cluster
  - name: staging-cluster
    namespaces:
      - production
      - critical
```

## Blocked Commands

The following commands are blocked on guarded contexts:

- `delete`
- `apply`
- `patch`
- `replace`
- `scale`
- `rollout`
- `drain`
- `cordon` / `uncordon`
- `taint`
- `label`
- `annotate`
- `edit`
- `set`

## Shell Alias (Recommended)

To always enable protection checks:

```bash
# ~/.bashrc or ~/.zshrc
alias k='kubectl guard exec --'
```

This allows `k delete pod nginx` to automatically check context protection.

## License

MIT
