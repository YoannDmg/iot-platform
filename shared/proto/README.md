# Protocol Buffers

Ce dossier contient les définitions Protocol Buffers pour la communication gRPC entre les microservices.

## 🔧 Installation des outils

### 1. Installer Protocol Buffers Compiler (protoc)

**macOS:**
```bash
brew install protobuf
```

**Linux (Debian/Ubuntu):**
```bash
sudo apt install -y protobuf-compiler
```

**Vérifier l'installation:**
```bash
protoc --version
# Doit afficher: libprotoc 3.x.x ou supérieur
```

### 2. Installer les plugins Go

```bash
# Plugin pour générer les structures Go
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Plugin pour générer le code gRPC
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

**Vérifier que les plugins sont dans le PATH:**
```bash
which protoc-gen-go
which protoc-gen-go-grpc
```

Si rien ne s'affiche, ajoute ceci à ton `~/.zshrc` ou `~/.bashrc`:
```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

## 📝 Générer le code

```bash
cd shared/proto
./generate.sh
```

Cela va créer les fichiers Go dans `shared/proto/device/` :
- `device.pb.go` : Structures de données
- `device_grpc.pb.go` : Code client/serveur gRPC

## 📁 Structure des fichiers

```
shared/proto/
├── device.proto          # Définitions des messages et services
├── generate.sh           # Script de génération
├── README.md            # Ce fichier
└── device/              # Code généré (créé automatiquement)
    ├── device.pb.go
    └── device_grpc.pb.go
```

## 🔄 Workflow

1. Modifier `device.proto`
2. Lancer `./generate.sh`
3. Le code Go est régénéré automatiquement
4. Utiliser les structures générées dans les services
