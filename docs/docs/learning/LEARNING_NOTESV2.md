---
id: LEARNING_NOTESV2
title: Notes d'apprentissage v2
sidebar_label: Notes v2
---

# Notes de Cours - Architecture Microservices IoT

> Document pédagogique évolutif - Un parcours d'apprentissage progressif sur la construction d'une plateforme IoT moderne

**Dernière mise à jour:** 2026-01-09
**Niveau:** Intermédiaire
**Prérequis:** Bases de Go, notions de réseaux et API

---

## Introduction

Ce document est conçu comme un cours universitaire progressif. Nous allons construire ensemble une plateforme IoT en utilisant une architecture microservices moderne. L'objectif n'est pas seulement d'apprendre à coder, mais de **comprendre pourquoi** nous faisons les choses d'une certaine manière.

### Philosophie d'apprentissage

Dans ce cours, nous suivons une approche pédagogique en trois temps:
1. **Comprendre le problème** - Pourquoi avons-nous besoin de cette solution?
2. **Explorer les concepts** - Comment cela fonctionne en théorie?
3. **Implémenter la solution** - Comment le faire concrètement?

---

## Table des matières

### Partie I - Fondements théoriques
1. [Qu'est-ce qu'une architecture microservices?](#1-quest-ce-quune-architecture-microservices)
2. [Communication entre services: comprendre les protocoles](#2-communication-entre-services-comprendre-les-protocoles)
3. [Les API modernes: REST, GraphQL et gRPC](#3-les-api-modernes-rest-graphql-et-grpc)

### Partie II - Concepts clés
4. [Protocol Buffers: le langage universel des services](#4-protocol-buffers-le-langage-universel-des-services)
5. [GraphQL: donner le contrôle au client](#5-graphql-donner-le-contrôle-au-client)
6. [La génération de code: travailler plus intelligemment](#6-la-génération-de-code-travailler-plus-intelligemment)

### Partie III - Mise en pratique
7. [Architecture de notre plateforme IoT](#7-architecture-de-notre-plateforme-iot)
8. [L'API Gateway: le portier intelligent](#8-lapi-gateway-le-portier-intelligent)
9. [Synchronisation et sécurité des données](#9-synchronisation-et-sécurité-des-données)

### Partie IV - Déploiement et bonnes pratiques
10. [Containerisation avec Docker](#10-containerisation-avec-docker)
11. [Patterns de conception et bonnes pratiques](#11-patterns-de-conception-et-bonnes-pratiques)
12. [Outils de développement et workflow](#12-outils-de-développement-et-workflow)

---

## Partie I - Fondements théoriques

### 1. Qu'est-ce qu'une architecture microservices?

#### 1.1 Le problème: l'application monolithique

Imaginez une grande entreprise où tout le monde travaille dans un seul bureau géant. Si une personne tombe malade, si l'électricité coupe, ou si quelqu'un fait une erreur, tout le monde est impacté. C'est exactement ce qui se passe avec une **application monolithique**.

```
┌─────────────────────────────────────────┐
│    APPLICATION MONOLITHIQUE             │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │  Interface Utilisateur          │   │
│  ├─────────────────────────────────┤   │
│  │  Logique Métier                 │   │
│  ├─────────────────────────────────┤   │
│  │  Accès aux Données              │   │
│  └─────────────────────────────────┘   │
│                                         │
│  Tout est dans un seul bloc             │
│  Déployé en une seule fois              │
│  Échec d'une partie = échec total       │
└─────────────────────────────────────────┘
```

**Problèmes concrets:**
- Un bug dans une petite fonctionnalité peut crasher toute l'application
- Impossible de mettre à jour une partie sans redéployer le tout
- L'équipe de développement doit se coordonner sur chaque changement
- Difficile de scaler: on doit dupliquer toute l'application même si seule une partie est sollicitée

#### 1.2 La solution: l'architecture microservices

Maintenant, imaginez que cette entreprise décide de créer des départements indépendants: comptabilité, ressources humaines, ventes, etc. Chaque département a son propre espace, ses propres outils, et communique avec les autres via des messages clairs (emails, appels, réunions).

```
┌──────────────────────────────────────────────────────────────┐
│                  ARCHITECTURE MICROSERVICES                   │
│                                                               │
│  ┌─────────────┐   ┌─────────────┐   ┌─────────────┐        │
│  │   Service   │   │   Service   │   │   Service   │        │
│  │     A       │◄──┤     B       │◄──┤     C       │        │
│  │             │   │             │   │             │        │
│  │  - DB A     │   │  - DB B     │   │  - DB C     │        │
│  └─────────────┘   └─────────────┘   └─────────────┘        │
│                                                               │
│  Chaque service est autonome                                 │
│  Déployé indépendamment                                      │
│  Échec d'un service ≠ échec total                            │
└──────────────────────────────────────────────────────────────┘
```

**Avantages:**
- **Isolation des pannes**: Un service qui échoue n'affecte pas les autres
- **Déploiement indépendant**: On peut mettre à jour le service A sans toucher au service B
- **Équipes autonomes**: Chaque équipe peut choisir sa technologie et son rythme
- **Scalabilité ciblée**: On peut dupliquer uniquement les services qui sont sollicités

#### 1.3 Le prix à payer

Comme toute solution, les microservices ont un coût:

**Complexité accrue:**
- Au lieu d'un seul programme, vous en avez plusieurs à gérer
- La communication entre services nécessite des protocoles et de la coordination
- Le débogage devient plus difficile (les erreurs peuvent traverser plusieurs services)

**Nouvelles questions à résoudre:**
- Comment les services se trouvent-ils les uns les autres?
- Comment gérer les transactions qui touchent plusieurs services?
- Comment monitorer l'ensemble du système?

**Quand utiliser des microservices?**

La règle d'or: **ne pas commencer par des microservices**. Commencez avec un monolithe, et quand vous ressentez la douleur de la taille (équipes trop grandes, déploiements trop longs, impossible de scaler une partie), alors passez aux microservices.

Dans notre cas (plateforme IoT), les microservices sont justifiés car:
- Nous avons des besoins de scalabilité très différents (collecter des données vs afficher un dashboard)
- Certaines parties nécessitent des technologies spécifiques (traitement temps réel en Rust)
- Les domaines métier sont clairement séparés (gestion des appareils, collecte de données, analytics)

---

### 2. Communication entre services: comprendre les protocoles

#### 2.1 Comment deux services se parlent?

Quand vous parlez à quelqu'un, vous utilisez un **langage** (français, anglais) et un **moyen de communication** (téléphone, email, face à face). C'est exactement pareil pour les services.

**Les trois dimensions de la communication:**

1. **Le transport** - Comment les bits voyagent sur le réseau
   - TCP/IP: comme un tuyau fiable qui garantit que tout arrive dans l'ordre
   - HTTP: un protocole au-dessus de TCP, comme une lettre avec une adresse et un contenu

2. **Le format de données** - Comment on structure l'information
   - JSON: texte lisible par les humains `{"name": "John", "age": 30}`
   - Protobuf: format binaire compact, rapide mais illisible directement

3. **Le pattern de communication** - Comment on organise l'échange
   - Request-Response: "Je pose une question, tu réponds" (synchrone)
   - Event-Driven: "Je publie une information, qui veut peut l'écouter" (asynchrone)

#### 2.2 HTTP: le langage du web

HTTP (HyperText Transfer Protocol) est le protocole que votre navigateur utilise pour charger cette page web. C'est un protocole **texte** basé sur le modèle request-response.

**Anatomie d'une requête HTTP:**

```
POST /api/devices HTTP/1.1          ← Méthode, chemin, version
Host: api.example.com               ← Headers (métadonnées)
Content-Type: application/json      ← Type de contenu
Authorization: Bearer token123

{                                   ← Body (corps)
  "name": "Sensor1",
  "type": "temperature"
}
```

**Anatomie d'une réponse HTTP:**

```
HTTP/1.1 201 Created                ← Code de statut
Content-Type: application/json
Location: /api/devices/abc123

{
  "id": "abc123",
  "name": "Sensor1",
  "type": "temperature",
  "created_at": "2026-01-09T10:00:00Z"
}
```

**Les codes de statut HTTP:**
- `2xx` = Succès (200 OK, 201 Created)
- `4xx` = Erreur du client (400 Bad Request, 404 Not Found)
- `5xx` = Erreur du serveur (500 Internal Server Error)

#### 2.3 HTTP/1.1 vs HTTP/2

**HTTP/1.1** (années 1990):
- Une requête à la fois par connexion TCP
- Headers en texte (verbeux et répétitif)
- Pas de compression des headers

**Analogie**: Imaginez un guichet de poste où vous devez faire la queue séparément pour chaque service, même si c'est le même employé qui traite toutes vos demandes.

**HTTP/2** (2015):
- **Multiplexing**: plusieurs requêtes en parallèle sur une seule connexion
- Headers compressés (HPACK)
- Server push (le serveur peut envoyer des données sans que le client demande)

**Analogie**: Maintenant vous donnez toutes vos demandes en une fois, l'employé les traite en parallèle, et vous rend tout d'un coup.

```
HTTP/1.1:
┌─────────┐                      ┌─────────┐
│ Client  │─── Request 1 ───────►│ Server  │
│         │◄── Response 1 ───────┤         │
│         │─── Request 2 ───────►│         │
│         │◄── Response 2 ───────┤         │
└─────────┘                      └─────────┘
   Séquentiel, lent

HTTP/2:
┌─────────┐                      ┌─────────┐
│ Client  │─── Request 1 ───────►│ Server  │
│         │─── Request 2 ───────►│         │
│         │─── Request 3 ───────►│         │
│         │◄── Response 2 ───────┤         │
│         │◄── Response 1 ───────┤         │
│         │◄── Response 3 ───────┤         │
└─────────┘                      └─────────┘
   Parallèle, rapide
```

**Pourquoi c'est important pour nous?**
gRPC utilise HTTP/2, ce qui le rend beaucoup plus efficace que REST classique, surtout quand on a beaucoup de petites requêtes.

---

### 3. Les API modernes: REST, GraphQL et gRPC

#### 3.1 REST: l'approche traditionnelle

REST (REpresentational State Transfer) est un style d'architecture qui utilise HTTP de manière standardisée.

**Les principes REST:**
- Les **ressources** sont identifiées par des URLs (`/devices/123`)
- On utilise les **verbes HTTP** pour les actions (GET, POST, PUT, DELETE)
- Les réponses utilisent des **codes de statut HTTP**
- Sans état (stateless): chaque requête contient toutes les informations nécessaires

**Exemple d'API REST pour gérer des appareils IoT:**

```
GET    /devices              → Liste tous les appareils
GET    /devices/123          → Obtient un appareil spécifique
POST   /devices              → Crée un nouvel appareil
PUT    /devices/123          → Met à jour un appareil
DELETE /devices/123          → Supprime un appareil
```

**Les limitations de REST:**

1. **Over-fetching** (trop de données):
   ```
   GET /devices/123
   Retourne: { id, name, type, status, metadata, history, logs, ... }

   Mais vous vouliez juste le nom!
   ```

2. **Under-fetching** (pas assez de données):
   ```
   GET /devices/123        → Obtient l'appareil
   GET /devices/123/user   → Obtient l'utilisateur propriétaire
   GET /users/456/profile  → Obtient le profil de l'utilisateur

   3 requêtes pour avoir toutes les infos!
   ```

3. **Manque de typage fort**:
   - La documentation peut être obsolète
   - Pas de validation automatique des requêtes
   - Le client doit deviner la structure des données

#### 3.2 GraphQL: le client prend le contrôle

GraphQL a été créé par Facebook en 2012 pour résoudre les problèmes de REST dans leur application mobile.

**L'idée révolutionnaire:**
Au lieu que le serveur décide quelles données envoyer, c'est le **client qui demande exactement ce dont il a besoin**.

**Analogie:**
- **REST**: Vous allez au restaurant et commandez un menu. Vous recevez entrée, plat, dessert même si vous vouliez juste le plat.
- **GraphQL**: Vous indiquez exactement ce que vous voulez: "juste le plat principal, sans les frites, et avec une sauce supplémentaire".

**Exemple concret:**

```graphql
# Le client envoie cette requête
query {
  device(id: "123") {
    name
    status
  }
}

# Le serveur répond avec EXACTEMENT ce qui a été demandé
{
  "data": {
    "device": {
      "name": "Temperature Sensor",
      "status": "ONLINE"
    }
  }
}
```

**Avantages:**
- Une seule requête pour obtenir des données de plusieurs ressources
- Le client reçoit exactement ce qu'il demande (ni plus, ni moins)
- Typage fort: le schéma définit exactement ce qui est possible
- Documentation auto-générée à partir du schéma

**Inconvénients:**
- Plus complexe à mettre en cache (pas d'URLs simples)
- Plus difficile à monitorer (toutes les requêtes vont au même endpoint)
- Peut être trop puissant: un client peut faire des requêtes très coûteuses

#### 3.3 gRPC: la performance avant tout

gRPC (gRPC Remote Procedure Call) a été créé par Google pour la communication entre leurs services internes.

**L'idée:**
Appeler une fonction sur un autre serveur comme si c'était une fonction locale, mais avec des garanties de performance et de typage.

**Analogie:**
Imaginez que vous avez un assistant dans une autre ville. Avec REST, vous lui envoyez des lettres (JSON). Avec gRPC, c'est comme si vous aviez un téléphone direct avec un langage secret ultra-compact que vous seuls comprenez.

**Caractéristiques:**
- Utilise **Protobuf** (format binaire) au lieu de JSON
- Utilise **HTTP/2** (multiplexing, streaming)
- **Typage fort obligatoire**: le contrat est défini dans un fichier `.proto`
- **Génération de code automatique** pour client et serveur

**Exemple de comparaison REST vs gRPC:**

```
REST (JSON):
{
  "id": "123",
  "name": "Temperature Sensor",
  "status": "online",
  "created_at": "2026-01-09T10:00:00Z"
}
→ ~150 bytes en JSON

gRPC (Protobuf):
[binaire incompréhensible pour les humains]
→ ~40 bytes en Protobuf

Performance: 3-4x plus rapide!
```

**Quand utiliser quoi?**

| Critère | REST | GraphQL | gRPC |
|---------|------|---------|------|
| **API publique** | ✅ Excellent | ✅ Excellent | ❌ Compliqué |
| **Communication interne** | ⚠️ Acceptable | ⚠️ Acceptable | ✅ Idéal |
| **Performance critique** | ❌ Moyen | ⚠️ Bon | ✅ Excellent |
| **Typage fort** | ❌ Faible | ✅ Fort | ✅ Très fort |
| **Streaming** | ❌ Compliqué | ⚠️ Possible | ✅ Native |
| **Cache HTTP** | ✅ Simple | ❌ Complexe | ❌ N/A |
| **Debugging** | ✅ Simple (JSON) | ✅ Simple (JSON) | ❌ Complexe (binaire) |

**Notre choix d'architecture:**
- **GraphQL** pour l'API publique (flexibilité pour les clients web/mobile)
- **gRPC** pour la communication entre services (performance et typage)

---

## Partie II - Concepts clés

### 4. Protocol Buffers: le langage universel des services

#### 4.1 Le problème de la communication inter-langages

Imaginez une entreprise internationale où l'équipe en France parle français, l'équipe au Japon parle japonais, et l'équipe aux USA parle anglais. Comment communiquent-ils efficacement?

**Solution classique (JSON):**
On utilise un langage simple (anglais/JSON), mais:
- C'est verbeux: beaucoup de mots pour peu d'information
- Pas de garantie: rien ne vous empêche d'écrire n'importe quoi
- Lent à parser: il faut interpréter le texte à chaque fois

**Solution moderne (Protocol Buffers):**
On crée un **dictionnaire partagé** qui définit exactement:
- Quels messages peuvent être envoyés
- Quelle est la structure de chaque message
- Quel est le type de chaque champ

#### 4.2 Comprendre Protocol Buffers par l'exemple

**Imaginez que vous envoyez une carte d'identité:**

Version JSON (texte):
```json
{
  "id": "12345",
  "name": "Alice Dupont",
  "age": 30,
  "email": "alice@example.com"
}
```
Taille: ~80 bytes

Version Protobuf (concept):
```
Définition du format:
- Champ 1: id (texte)
- Champ 2: name (texte)
- Champ 3: age (nombre)
- Champ 4: email (texte)

Message encodé:
[1|12345][2|Alice Dupont][3|30][4|alice@example.com]
```
Taille: ~40 bytes (compression binaire)

#### 4.3 Anatomie d'un fichier .proto

Prenons un exemple concret pour notre plateforme IoT:

```protobuf
// shared/proto/device.proto

// Version du protocole (toujours proto3 maintenant)
syntax = "proto3";

// Nom du package (comme un namespace)
package device;

// Option pour la génération Go
option go_package = "github.com/yourusername/iot-platform/shared/proto";

// ============================================
// DÉFINITION D'UN MESSAGE (= une structure)
// ============================================

message Device {
  // Format: type nom = numéro_unique;

  string id = 1;          // Identifiant unique
  string name = 2;        // Nom de l'appareil
  string type = 3;        // Type (sensor, actuator, etc.)
  DeviceStatus status = 4; // Statut (utilise l'enum ci-dessous)
  int64 created_at = 5;   // Timestamp de création
  int64 last_seen = 6;    // Dernière activité

  // map = dictionnaire clé-valeur
  map<string, string> metadata = 7;
}

// ============================================
// ÉNUMÉRATION (= liste de valeurs possibles)
// ============================================

enum DeviceStatus {
  // IMPORTANT: 0 est toujours la valeur par défaut
  UNKNOWN = 0;
  ONLINE = 1;
  OFFLINE = 2;
  ERROR = 3;
  MAINTENANCE = 4;
}

// ============================================
// DÉFINITION D'UN SERVICE (= API)
// ============================================

service DeviceService {
  // Définit les fonctions que le service expose
  // Format: rpc NomFonction(TypeRequête) returns (TypeRéponse);

  rpc CreateDevice(CreateDeviceRequest) returns (CreateDeviceResponse);
  rpc GetDevice(GetDeviceRequest) returns (GetDeviceResponse);
  rpc ListDevices(ListDevicesRequest) returns (ListDevicesResponse);
  rpc UpdateDevice(UpdateDeviceRequest) returns (UpdateDeviceResponse);
  rpc DeleteDevice(DeleteDeviceRequest) returns (DeleteDeviceResponse);

  // Streaming: le serveur envoie plusieurs messages
  rpc WatchDevices(WatchDevicesRequest) returns (stream Device);
}

// ============================================
// MESSAGES POUR LES REQUÊTES/RÉPONSES
// ============================================

message CreateDeviceRequest {
  string name = 1;
  string type = 2;
  map<string, string> metadata = 3;
}

message CreateDeviceResponse {
  Device device = 1;  // Le device créé
}

message GetDeviceRequest {
  string id = 1;
}

message GetDeviceResponse {
  Device device = 1;
}

// ... autres messages
```

#### 4.4 Les règles d'or de Protobuf

**Règle 1: Les numéros de champ sont sacrés**

```protobuf
message Device {
  string id = 1;      // Ce numéro ne doit JAMAIS changer
  string name = 2;    // Ce numéro ne doit JAMAIS changer
}
```

Pourquoi? Parce que dans le format binaire, on utilise ces numéros, pas les noms. Si vous changez le numéro, c'est comme si vous changiez le champ lui-même.

**OK: Ajouter un nouveau champ**
```protobuf
message Device {
  string id = 1;
  string name = 2;
  string description = 3;  // Nouveau champ
}
```

**OK: Supprimer un champ (marquer comme réservé)**
```protobuf
message Device {
  string id = 1;
  // string name = 2;  // On ne l'utilise plus
  reserved 2;           // Empêche la réutilisation du numéro
  string description = 3;
}
```

**INTERDIT: Changer le numéro d'un champ**
```protobuf
// Version 1
message Device {
  string id = 1;
  string name = 2;
}

// Version 2 - ❌ ERREUR!
message Device {
  string id = 1;
  string name = 3;  // Le changement du numéro casse la compatibilité!
}
```

**Règle 2: Le champ 0 d'un enum est la valeur par défaut**

```protobuf
enum DeviceStatus {
  UNKNOWN = 0;      // Si aucune valeur n'est spécifiée, c'est UNKNOWN
  ONLINE = 1;
  OFFLINE = 2;
}
```

#### 4.5 Comment fonctionne la génération de code?

Quand vous exécutez `protoc` (le compilateur Protocol Buffers), voici ce qui se passe:

```
┌─────────────────┐
│  device.proto   │  ← Votre définition (source de vérité)
└────────┬────────┘
         │
         │ protoc --go_out=. --go-grpc_out=.
         │
    ┌────▼──────────────────────────────────┐
    │                                       │
    ▼                                       ▼
┌────────────────┐                 ┌────────────────────┐
│ device.pb.go   │                 │ device_grpc.pb.go  │
│                │                 │                    │
│ • Structs      │                 │ • Interfaces       │
│ • Enums        │                 │ • Client           │
│ • Serialization│                 │ • Server stub      │
└────────────────┘                 └────────────────────┘
```

**Exemple de code généré:**

Protobuf:
```protobuf
message Device {
  string id = 1;
  string name = 2;
  DeviceStatus status = 3;
}
```

Code Go généré (simplifié):
```go
// device.pb.go

type Device struct {
    Id     string       `protobuf:"bytes,1,opt,name=id,proto3"`
    Name   string       `protobuf:"bytes,2,opt,name=name,proto3"`
    Status DeviceStatus `protobuf:"varint,3,opt,name=status,proto3"`
}

// Méthodes de serialization/deserialization générées automatiquement
func (m *Device) Marshal() ([]byte, error) { /* ... */ }
func (m *Device) Unmarshal(data []byte) error { /* ... */ }
```

Service:
```protobuf
service DeviceService {
  rpc CreateDevice(CreateDeviceRequest) returns (CreateDeviceResponse);
}
```

Code Go généré (simplifié):
```go
// device_grpc.pb.go

// Interface que VOUS devez implémenter (serveur)
type DeviceServiceServer interface {
    CreateDevice(context.Context, *CreateDeviceRequest) (*CreateDeviceResponse, error)
}

// Client généré automatiquement (pour appeler le service)
type DeviceServiceClient interface {
    CreateDevice(ctx context.Context, in *CreateDeviceRequest, opts ...grpc.CallOption) (*CreateDeviceResponse, error)
}

// Fonction pour créer un client
func NewDeviceServiceClient(conn *grpc.ClientConn) DeviceServiceClient {
    return &deviceServiceClient{conn}
}
```

#### 4.6 Pourquoi Protocol Buffers est révolutionnaire

**Avantage 1: Contrat explicite**
- Le fichier `.proto` EST la documentation
- Si le serveur change son API sans modifier le `.proto`, la compilation échoue
- Pas de surprise à l'exécution

**Avantage 2: Multi-langage**
- Définissez une fois en `.proto`
- Générez du code en Go, Java, Python, C++, JavaScript...
- Tous les services parlent le même "langage"

**Avantage 3: Performance**
- Format binaire compact
- Parsing ultra-rapide (pas de JSON parsing)
- Parfait pour les systèmes haute performance

**Avantage 4: Évolution compatible**
- Ajoutez des champs sans casser les anciens clients
- Vieux et nouveaux services peuvent communiquer

---

### 5. GraphQL: donner le contrôle au client

#### 5.1 La philosophie GraphQL

GraphQL n'est pas juste une alternative technique à REST. C'est une **philosophie différente** de la conception d'API.

**REST dit:** "Je suis le serveur, je décide quelles données je vous donne"
**GraphQL dit:** "Tu es le client, dis-moi exactement ce dont tu as besoin"

#### 5.2 Les trois concepts fondamentaux

**1. Le schéma: le contrat**

Le schéma GraphQL définit:
- Quels **types** de données existent
- Quelles **requêtes** (lectures) sont possibles
- Quelles **mutations** (écritures) sont possibles
- Quelles **souscriptions** (temps réel) sont possibles

```graphql
# Définition d'un type
type Device {
  id: ID!           # ID! signifie "ID obligatoire (non null)"
  name: String!
  type: String!
  status: DeviceStatus!
  metadata: [MetadataEntry!]!  # Liste d'entrées (non null)
}

# Énumération
enum DeviceStatus {
  UNKNOWN
  ONLINE
  OFFLINE
  ERROR
  MAINTENANCE
}

# Type pour les métadonnées
type MetadataEntry {
  key: String!
  value: String!
}

# Types d'input (pour les mutations)
input CreateDeviceInput {
  name: String!
  type: String!
  metadata: [MetadataEntryInput!]
}

# Requêtes (lecture)
type Query {
  device(id: ID!): Device
  devices(page: Int!, pageSize: Int!): DeviceConnection!
}

# Mutations (écriture)
type Mutation {
  createDevice(input: CreateDeviceInput!): Device!
  updateDevice(id: ID!, input: UpdateDeviceInput!): Device!
  deleteDevice(id: ID!): Boolean!
}

# Souscriptions (temps réel)
type Subscription {
  deviceUpdated: Device!
}
```

**2. Les requêtes: demander exactement ce qu'on veut**

```graphql
# Exemple 1: Juste le minimum
query {
  device(id: "123") {
    id
    name
  }
}

# Réponse:
{
  "data": {
    "device": {
      "id": "123",
      "name": "Temperature Sensor"
    }
  }
}

# Exemple 2: Plus de détails
query {
  device(id: "123") {
    id
    name
    status
    metadata {
      key
      value
    }
  }
}

# Exemple 3: Plusieurs requêtes en une seule
query {
  device1: device(id: "123") {
    name
  }
  device2: device(id: "456") {
    name
  }
  allDevices: devices(page: 1, pageSize: 10) {
    devices {
      id
      name
    }
    total
  }
}
```

**3. Les mutations: modifier les données**

```graphql
mutation CreateNewDevice {
  createDevice(input: {
    name: "Temperature Sensor Living Room"
    type: "temperature_sensor"
    metadata: [
      { key: "location", value: "living_room" }
      { key: "floor", value: "1" }
      { key: "building", value: "A" }
    ]
  }) {
    id        # On demande l'ID du device créé
    name
    status
    createdAt
  }
}

# Réponse:
{
  "data": {
    "createDevice": {
      "id": "abc123xyz",
      "name": "Temperature Sensor Living Room",
      "status": "ONLINE",
      "createdAt": 1704801600
    }
  }
}
```

#### 5.3 Les avantages de GraphQL expliqués par des exemples réels

**Scénario 1: Application mobile avec connexion lente**

Avec REST:
```
GET /devices/123          → 150 KB (toutes les infos)
Temps: 2 secondes
```

Avec GraphQL:
```graphql
query {
  device(id: "123") {
    name      # Juste ce qui sera affiché
    status
  }
}
→ 500 bytes
Temps: 0.1 seconde
```

**Scénario 2: Affichage d'un dashboard complexe**

Avec REST (N+1 problem):
```
1. GET /devices              → Liste des devices
2. GET /devices/1/location   → Location du device 1
3. GET /devices/2/location   → Location du device 2
4. GET /devices/3/location   → Location du device 3
...
100 requêtes pour 100 devices!
```

Avec GraphQL:
```graphql
query {
  devices(page: 1, pageSize: 100) {
    devices {
      id
      name
      location {    # Relations chargées en une seule requête
        building
        floor
        room
      }
    }
  }
}
# 1 seule requête!
```

**Scénario 3: Évolution de l'API**

Version 1 de votre app mobile affiche juste le nom.
Version 2 affiche aussi le statut.
Version 3 affiche aussi les métadonnées.

Avec REST:
- Créer `/devices/v1`, `/devices/v2`, `/devices/v3`
- Ou ajouter des paramètres: `/devices?fields=name,status`
- Ou utiliser des headers: `Accept: application/vnd.api.v2+json`

Avec GraphQL:
- Le schéma reste le même
- Chaque version de l'app demande ce qu'elle veut:

```graphql
# App v1
query { device(id: "123") { name } }

# App v2
query { device(id: "123") { name status } }

# App v3
query { device(id: "123") { name status metadata { key value } } }
```

Pas besoin de versions!

#### 5.4 Les pièges de GraphQL

**Piège 1: Requêtes complexes et coûteuses**

Un client malveillant (ou mal conçu) peut faire:

```graphql
query {
  devices(page: 1, pageSize: 10000) {  # Demande 10000 devices
    devices {
      id
      name
      metadata { key value }  # Pour chacun, charge les métadonnées
      history {               # Et l'historique complet
        timestamp
        value
      }
    }
  }
}
```

Cette requête pourrait exploser votre base de données!

**Solutions:**
- Limiter la profondeur des requêtes (max 5 niveaux)
- Limiter la complexité (coût calculé par champ)
- Pagination obligatoire
- Timeout sur les requêtes

**Piège 2: Cache HTTP compliqué**

Avec REST:
```
GET /devices/123  → Cache avec l'URL comme clé
```

Avec GraphQL:
```
POST /query
Body: { "query": "query { device(id: \"123\") { name } }" }
```

Toutes les requêtes vont au même endpoint (`/query`), donc difficile à cacher avec un cache HTTP classique. Il faut utiliser un cache applicatif.

---

### 6. La génération de code: travailler plus intelligemment

#### 6.1 Le principe DRY: Don't Repeat Yourself

Imaginez que vous écrivez un livre et que quelqu'un d'autre écrit un résumé de votre livre. Si vous modifiez un chapitre, le résumé devient obsolète. C'est le même problème avec le code.

**Sans génération:**
```
1. Vous écrivez le schéma API (ce qui est possible)
2. Vous écrivez les structs Go (les types de données)
3. Vous écrivez le code de serialization
4. Vous écrivez le client
5. Vous écrivez le serveur

Si vous changez quelque chose, vous devez modifier 5 endroits!
Risque d'oubli = bugs
```

**Avec génération:**
```
1. Vous écrivez le schéma API (source de vérité)
2. Le code est généré automatiquement

Si vous changez le schéma:
- Lancer la génération
- Si le code dépendant n'est pas mis à jour, la compilation échoue
- Impossible d'oublier quelque chose!
```

#### 6.2 Notre pipeline de génération

```
┌─────────────────────────────────────────────────┐
│                SOURCES DE VÉRITÉ                │
│  (Ce qu'on écrit à la main)                     │
├─────────────────────────────────────────────────┤
│                                                 │
│  shared/proto/device.proto                      │
│  └─→ Définit les messages et services gRPC      │
│                                                 │
│  services/api-gateway/graph/schema.graphql      │
│  └─→ Définit l'API GraphQL publique             │
│                                                 │
└─────────────────────────────────────────────────┘
                      │
                      │ make generate
                      │
        ┌─────────────┴─────────────┐
        │                           │
        ▼                           ▼
┌───────────────────┐     ┌──────────────────────┐
│  GÉNÉRATION gRPC  │     │ GÉNÉRATION GraphQL   │
├───────────────────┤     ├──────────────────────┤
│                   │     │                      │
│ protoc            │     │ gqlgen               │
│                   │     │                      │
│ Génère:           │     │ Génère:              │
│ • device.pb.go    │     │ • generated.go       │
│ • device_grpc.pb  │     │ • models_gen.go      │
│   .go             │     │ • schema.resolvers   │
│                   │     │   .go (stubs)        │
└───────────────────┘     └──────────────────────┘
                      │
                      ▼
        ┌─────────────────────────┐
        │   CODE IMPLÉMENTÉ       │
        │   (Ce qu'on écrit)      │
        ├─────────────────────────┤
        │                         │
        │ • resolvers_impl.go     │
        │   → Logique GraphQL     │
        │                         │
        │ • server.go             │
        │   → Logique gRPC        │
        │                         │
        └─────────────────────────┘
```

#### 6.3 Workflow de développement

**Scénario: Ajouter une nouvelle fonctionnalité "Obtenir les statistiques des devices"**

**Étape 1: Modifier le schéma Proto**

```protobuf
// shared/proto/device.proto

message DeviceStats {
  int32 total_devices = 1;
  int32 online_devices = 2;
  int32 offline_devices = 3;
  int32 error_devices = 4;
}

message GetStatsRequest {}

message GetStatsResponse {
  DeviceStats stats = 1;
}

service DeviceService {
  // Ajouter la nouvelle méthode
  rpc GetStats(GetStatsRequest) returns (GetStatsResponse);
  // ... autres méthodes
}
```

**Étape 2: Générer le code gRPC**

```bash
make generate-proto
```

Résultat: Le code Go est généré automatiquement.

**Étape 3: Implémenter la méthode dans le serveur**

```go
// services/device-manager/server.go

func (s *DeviceServer) GetStats(
    ctx context.Context,
    req *pb.GetStatsRequest,
) (*pb.GetStatsResponse, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    stats := &pb.DeviceStats{
        TotalDevices: int32(len(s.devices)),
    }

    for _, device := range s.devices {
        switch device.Status {
        case pb.DeviceStatus_ONLINE:
            stats.OnlineDevices++
        case pb.DeviceStatus_OFFLINE:
            stats.OfflineDevices++
        case pb.DeviceStatus_ERROR:
            stats.ErrorDevices++
        }
    }

    return &pb.GetStatsResponse{Stats: stats}, nil
}
```

**Étape 4: Modifier le schéma GraphQL**

```graphql
# services/api-gateway/graph/schema.graphql

type DeviceStats {
  totalDevices: Int!
  onlineDevices: Int!
  offlineDevices: Int!
  errorDevices: Int!
}

type Query {
  # Ajouter la nouvelle query
  stats: DeviceStats!
  # ... autres queries
}
```

**Étape 5: Générer le code GraphQL**

```bash
make generate-graphql
```

Résultat: Un stub est créé dans `schema.resolvers.go`.

**Étape 6: Implémenter le resolver**

```go
// services/api-gateway/graph/resolvers_impl.go

func (r *Resolver) StatsImpl(ctx context.Context) (*model.DeviceStats, error) {
    // Appeler le service gRPC
    resp, err := r.DeviceClient.GetStats(ctx, &pb.GetStatsRequest{})
    if err != nil {
        return nil, err
    }

    // Convertir Protobuf → GraphQL
    return &model.DeviceStats{
        TotalDevices:  int(resp.Stats.TotalDevices),
        OnlineDevices: int(resp.Stats.OnlineDevices),
        OfflineDevices: int(resp.Stats.OfflineDevices),
        ErrorDevices:  int(resp.Stats.ErrorDevices),
    }, nil
}
```

**Étape 7: Tester**

```bash
# Compiler (si erreur = incompatibilité de types)
go build

# Lancer les services
make dev

# Tester dans GraphQL Playground
query {
  stats {
    totalDevices
    onlineDevices
    offlineDevices
  }
}
```

#### 6.4 Les avantages de cette approche

**1. Sécurité à la compilation**

Si vous oubliez d'implémenter une méthode, le code ne compile pas:

```go
// Erreur de compilation:
*Resolver does not implement generated.QueryResolver
(missing method Stats)
```

**2. Refactoring sûr**

Si vous renommez un champ dans le schéma et régénérez:

```protobuf
// Avant:
string device_name = 2;

// Après:
string name = 2;  // Renommé
```

Tous les endroits où vous utilisiez `device_name` génèreront une erreur de compilation. Vous ne pouvez pas oublier de mettre à jour.

**3. Documentation toujours à jour**

Le schéma EST la documentation. Si la documentation n'est pas à jour, c'est que le schéma ne l'est pas, et donc le code ne compile pas.

---

## Partie III - Mise en pratique

### 7. Architecture de notre plateforme IoT

#### 7.1 Vue d'ensemble du système

Notre plateforme IoT est conçue pour gérer des milliers d'appareils connectés. Voici comment elle est organisée:

```
┌──────────────────────────────────────────────────────────────┐
│                         CLIENTS                               │
│                                                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │ Web App     │  │ Mobile App  │  │ IoT Device  │          │
│  │ (React)     │  │ (Flutter)   │  │ (ESP32)     │          │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘          │
│         │                 │                 │                 │
└─────────┼─────────────────┼─────────────────┼─────────────────┘
          │                 │                 │
          │ HTTP/GraphQL    │                 │ MQTT
          │                 │                 │
┌─────────▼─────────────────▼─────────────────┼─────────────────┐
│                    GATEWAY LAYER             │                 │
│                                              │                 │
│  ┌────────────────────────────────┐          │                 │
│  │      API Gateway               │          │                 │
│  │      (Port 8080)               │          │                 │
│  │                                │          │                 │
│  │  • GraphQL Server              │          │                 │
│  │  • Authentication              │          │                 │
│  │  • Rate Limiting               │          │                 │
│  └────────────┬───────────────────┘          │                 │
│               │                              │                 │
└───────────────┼──────────────────────────────┼─────────────────┘
                │                              │
                │ gRPC                         │
                │                              │
┌───────────────┼──────────────────────────────┼─────────────────┐
│          SERVICE LAYER (gRPC)                │                 │
│               │                              │                 │
│  ┌────────────▼────────────┐   ┌────────────▼─────────────┐   │
│  │   Device Manager        │   │   Data Collector         │   │
│  │   (Port 8081)           │   │   (Port 8082)            │   │
│  │                         │   │                          │   │
│  │  • CRUD operations      │   │  • Receive telemetry     │   │
│  │  • Device status        │   │  • Process streams       │   │
│  │  • Metadata mgmt        │   │  • Aggregate data        │   │
│  └────────────┬────────────┘   └────────────┬─────────────┘   │
│               │                              │                 │
└───────────────┼──────────────────────────────┼─────────────────┘
                │                              │
┌───────────────▼──────────────────────────────▼─────────────────┐
│                    DATA LAYER                                  │
│                                                                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐        │
│  │ PostgreSQL   │  │    Redis     │  │  TimescaleDB │        │
│  │              │  │              │  │              │        │
│  │ • Devices    │  │ • Cache      │  │ • Telemetry  │        │
│  │ • Users      │  │ • Sessions   │  │ • Metrics    │        │
│  │ • Config     │  │ • Pub/Sub    │  │ • Analytics  │        │
│  └──────────────┘  └──────────────┘  └──────────────┘        │
└────────────────────────────────────────────────────────────────┘
```

#### 7.2 Pourquoi cette architecture en couches?

**Couche 1: Clients (frontière externe)**
- **Responsabilité**: Interface utilisateur, expérience utilisateur
- **Protocole**: HTTP/GraphQL (flexible, facile à utiliser)
- **Exemple**: "Afficher la liste des devices avec leur statut"

**Couche 2: Gateway (point d'entrée)**
- **Responsabilité**: Validation, authentification, orchestration
- **Analogie**: Le réceptionniste d'un hôtel qui vérifie votre identité et vous dirige vers le bon service
- **Pourquoi**: Protéger les services internes, offrir une API uniforme

**Couche 3: Services (logique métier)**
- **Responsabilité**: Exécuter la logique métier, gérer les données
- **Protocole**: gRPC (rapide, typé, efficace)
- **Exemple**: "Créer un device, valider les données, le stocker en base"

**Couche 4: Données (persistance)**
- **Responsabilité**: Stocker et retrouver les données
- **Différents types**: PostgreSQL (relationnel), Redis (cache), TimescaleDB (séries temporelles)

#### 7.3 Le flux de données expliqué

**Scénario: Un utilisateur crée un nouveau capteur de température**

```
┌─────────────────────────────────────────────────────────────────┐
│ ÉTAPE 1: L'utilisateur remplit un formulaire                    │
└─────────────────────────────────────────────────────────────────┘
   Navigateur Web
   └─→ Form: Name="Capteur Salon", Type="Temperature"

┌─────────────────────────────────────────────────────────────────┐
│ ÉTAPE 2: Le frontend envoie une mutation GraphQL                │
└─────────────────────────────────────────────────────────────────┘
   POST http://api.example.com:8080/query
   {
     "query": "mutation {
       createDevice(input: {
         name: \"Capteur Salon\",
         type: \"temperature_sensor\"
       }) { id name status }
     }"
   }

┌─────────────────────────────────────────────────────────────────┐
│ ÉTAPE 3: API Gateway reçoit et valide                           │
└─────────────────────────────────────────────────────────────────┘
   API Gateway (Port 8080)
   1. Parse la requête GraphQL
   2. Valide la syntaxe
   3. Vérifie l'authentification
   4. Appelle le resolver CreateDevice

┌─────────────────────────────────────────────────────────────────┐
│ ÉTAPE 4: Resolver convertit GraphQL → gRPC                      │
└─────────────────────────────────────────────────────────────────┘
   Resolver.CreateDevice()
   └─→ Crée une requête Protobuf:
       CreateDeviceRequest{
         Name: "Capteur Salon",
         Type: "temperature_sensor",
       }

┌─────────────────────────────────────────────────────────────────┐
│ ÉTAPE 5: Appel gRPC vers Device Manager                         │
└─────────────────────────────────────────────────────────────────┘
   gRPC call: localhost:8081
   Protocole: HTTP/2 + Protobuf (binaire)

┌─────────────────────────────────────────────────────────────────┐
│ ÉTAPE 6: Device Manager traite la logique métier                │
└─────────────────────────────────────────────────────────────────┘
   DeviceServer.CreateDevice()
   1. Valide les données (name non vide, type valide)
   2. Génère un UUID: "abc123xyz"
   3. Crée l'objet Device:
      Device{
        Id: "abc123xyz",
        Name: "Capteur Salon",
        Type: "temperature_sensor",
        Status: ONLINE,
        CreatedAt: 1704801600,
      }
   4. Stocke en base de données (ou map en mémoire pour le moment)
   5. Retourne la réponse Protobuf

┌─────────────────────────────────────────────────────────────────┐
│ ÉTAPE 7: Resolver convertit gRPC → GraphQL                      │
└─────────────────────────────────────────────────────────────────┘
   Resolver reçoit la réponse Protobuf
   └─→ Convertit en type GraphQL:
       {
         "id": "abc123xyz",
         "name": "Capteur Salon",
         "status": "ONLINE"
       }

┌─────────────────────────────────────────────────────────────────┐
│ ÉTAPE 8: API Gateway retourne JSON au client                    │
└─────────────────────────────────────────────────────────────────┘
   HTTP/1.1 200 OK
   Content-Type: application/json
   {
     "data": {
       "createDevice": {
         "id": "abc123xyz",
         "name": "Capteur Salon",
         "status": "ONLINE"
       }
     }
   }

┌─────────────────────────────────────────────────────────────────┐
│ ÉTAPE 9: Le frontend affiche le résultat                        │
└─────────────────────────────────────────────────────────────────┘
   "✓ Capteur créé avec succès!"
```

**Résumé du parcours:**
```
Web App → GraphQL (JSON) → API Gateway → gRPC (Protobuf) → Device Manager → Database
        ←               ←              ←                 ←                ←
```

#### 7.4 Décisions architecturales et justifications

**Décision 1: GraphQL pour l'API publique**

**Pourquoi?**
- Les applications web/mobile ont des besoins très différents
- Un téléphone avec 3G ne veut pas charger 10 KB quand 500 bytes suffisent
- Les développeurs frontend veulent l'autonomie
- Documentation auto-générée (GraphQL Playground)

**Alternative rejetée:** REST avec différentes versions/endpoints
**Raison du rejet:** Explosion du nombre d'endpoints, maintenance difficile

---

**Décision 2: gRPC pour la communication interne**

**Pourquoi?**
- Performance critique quand on a des milliers de requêtes/seconde
- Typage fort évite les erreurs de communication
- HTTP/2 permet le multiplexing (plusieurs requêtes en parallèle)
- Streaming natif pour les données temps réel

**Alternative rejetée:** REST/JSON entre services
**Raison du rejet:** Trop lent, pas de typage fort, pas de streaming efficace

---

**Décision 3: Séparation Device Manager / Data Collector**

**Pourquoi?**
- Besoins de scalabilité différents:
  - Device Manager: peu de requêtes, CRUD simple
  - Data Collector: des milliers de messages/seconde, traitement streaming
- Technologies différentes:
  - Device Manager: Go (simplicité, productivité)
  - Data Collector: Rust (performance, sécurité mémoire)
- Pannes indépendantes: Si Data Collector crashe, on peut toujours gérer les devices

---

**Décision 4: Bases de données différentes selon les besoins**

**PostgreSQL (relationnel):**
- Pour: Devices, Users, Configuration
- Pourquoi: Relations entre entités, transactions ACID, requêtes complexes

**Redis (cache clé-valeur):**
- Pour: Sessions, cache de requêtes, pub/sub
- Pourquoi: Ultra-rapide (en mémoire), TTL automatique

**TimescaleDB (séries temporelles):**
- Pour: Données de télémétrie (température, humidité, etc.)
- Pourquoi: Optimisé pour les données avec timestamp, agrégations rapides

---

### 8. L'API Gateway: le portier intelligent

#### 8.1 Le rôle de l'API Gateway

L'API Gateway est comme le **réceptionniste d'un grand immeuble**. Tous les visiteurs (clients) passent par lui, et il:
- Vérifie leur identité (authentification)
- Dirige vers le bon département (routage)
- Traduit les langues (GraphQL ↔ gRPC)
- Limite l'accès (rate limiting)
- Journalise les visites (logging)

```
                    ┌──────────────────────┐
                    │   API GATEWAY        │
                    │   (Port 8080)        │
                    ├──────────────────────┤
                    │                      │
Clients →  HTTP     │  ┌────────────────┐  │
(Web/Mobile)  →     │  │ GraphQL Server │  │
           GraphQL  │  └───────┬────────┘  │
           JSON     │          │           │
                    │  ┌───────▼────────┐  │
                    │  │   Resolvers    │  │  ← Logique d'orchestration
                    │  │   (Impl)       │  │
                    │  └───────┬────────┘  │
                    │          │           │
                    │  ┌───────▼────────┐  │
                    │  │  gRPC Clients  │  │  ← Connexions persistantes
                    │  │                │  │
                    │  │ • DeviceClient │  │
                    │  │ • UserClient   │  │
                    │  │ • DataClient   │  │
                    │  └───────┬────────┘  │
                    └──────────┼───────────┘
                               │
                               │ gRPC (HTTP/2 + Protobuf)
                    ┌──────────▼───────────┐
                    │   Services internes  │
                    │   (Microservices)    │
                    └──────────────────────┘
```

#### 8.2 Organisation du code

Voici comment le code de l'API Gateway est structuré:

```
services/api-gateway/
│
├── main.go                          ← Point d'entrée
│   └─→ Initialise tout et lance le serveur
│
├── graph/
│   ├── schema.graphql               ← SOURCE DE VÉRITÉ
│   │   └─→ Définit l'API publique (queries, mutations, types)
│   │
│   ├── resolver.go                  ← Structure Resolver
│   │   └─→ Contient les dépendances (clients gRPC)
│   │
│   ├── schema.resolvers.go          ← Stubs GÉNÉRÉS
│   │   └─→ Fonctions qui délèguent aux *Impl
│   │
│   ├── resolvers_impl.go            ← Implémentations MANUELLES
│   │   └─→ Logique réelle (GraphQL → gRPC → GraphQL)
│   │
│   ├── generated/
│   │   └── generated.go             ← Serveur GraphQL GÉNÉRÉ
│   │
│   └── model/
│       └── models_gen.go            ← Types GraphQL GÉNÉRÉS
│
├── grpc/
│   └── client.go                    ← Wrapper pour les clients gRPC
│       └─→ Gère les connexions persistantes
│
├── gqlgen.yml                       ← Configuration de gqlgen
└── go.mod                           ← Dépendances Go
```

**Séparation des responsabilités:**

| Fichier | Généré? | Rôle | On le modifie? |
|---------|---------|------|----------------|
| `schema.graphql` | ❌ Non | Définit l'API | ✅ Oui |
| `resolver.go` | ❌ Non | Structure avec dépendances | ✅ Oui (injection) |
| `schema.resolvers.go` | ✅ Oui | Stubs qui délèguent | ⚠️ Parfois (délégation) |
| `resolvers_impl.go` | ❌ Non | Logique métier | ✅ Oui (souvent) |
| `generated.go` | ✅ Oui | Serveur GraphQL | ❌ Jamais |
| `models_gen.go` | ✅ Oui | Types Go | ❌ Jamais |
| `grpc/client.go` | ❌ Non | Client gRPC | ✅ Oui (ajout services) |

#### 8.3 Comprendre le client gRPC persistant

**Le problème:**

Créer une connexion gRPC prend du temps (handshake TCP, négociation HTTP/2). Si on crée une nouvelle connexion pour chaque requête GraphQL:

```go
// ❌ INEFFICACE
func CreateDevice(...) {
    conn, _ := grpc.Dial("localhost:8081", ...)  // 50ms de latence!
    client := pb.NewDeviceServiceClient(conn)
    resp, _ := client.CreateDevice(...)
    conn.Close()
}
```

Chaque requête GraphQL prend 50ms+ juste pour la connexion!

**La solution: connexion persistante**

```go
// ✅ EFFICACE
// Au démarrage du serveur, créer UNE SEULE connexion
deviceClient, _ := grpc.NewClient("localhost:8081", ...)

// Chaque requête réutilise la connexion
func CreateDevice(...) {
    resp, _ := deviceClient.CreateDevice(...)  // < 1ms
}
```

**Implémentation concrète:**

```go
// services/api-gateway/grpc/client.go

package grpc

import (
    "context"
    "fmt"
    "log"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    pb "github.com/yourusername/iot-platform/shared/proto"
)

// DeviceClient encapsule la connexion gRPC
type DeviceClient struct {
    conn   *grpc.ClientConn           // Connexion TCP/HTTP2 persistante
    client pb.DeviceServiceClient     // Client généré
}

// NewDeviceClient crée et initialise le client
func NewDeviceClient(address string) (*DeviceClient, error) {
    // Context avec timeout pour éviter de bloquer indéfiniment
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // CHANGEMENT RÉCENT: grpc.DialContext → grpc.NewClient
    // Raison: Linting Go 1.23+, DialContext est deprecated
    conn, err := grpc.NewClient(
        address,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to connect: %w", err)
    }

    // Vérifier que la connexion fonctionne
    log.Printf("✅ Connected to Device Manager at %s", address)

    return &DeviceClient{
        conn:   conn,
        client: pb.NewDeviceServiceClient(conn),
    }, nil
}

// Close ferme proprement la connexion
func (c *DeviceClient) Close() error {
    if c.conn != nil {
        return c.conn.Close()
    }
    return nil
}

// GetClient retourne le client gRPC pour l'utiliser
func (c *DeviceClient) GetClient() pb.DeviceServiceClient {
    return c.client
}
```

**Note importante:** Nous avons récemment migré de `grpc.DialContext` vers `grpc.NewClient` suite aux recommandations du linter Go. C'est la nouvelle manière recommandée de créer des connexions gRPC.

#### 8.4 Injection de dépendances dans les resolvers

**Principe:**

Au lieu que chaque resolver crée ses propres clients, on **injecte** les dépendances dans une structure partagée.

```go
// services/api-gateway/graph/resolver.go

package graph

import (
    pb "github.com/yourusername/iot-platform/shared/proto"
)

// Resolver contient toutes les dépendances nécessaires
// Cette structure est créée UNE FOIS au démarrage
type Resolver struct {
    // Clients gRPC (connexions persistantes)
    DeviceClient pb.DeviceServiceClient

    // On pourra ajouter d'autres dépendances plus tard:
    // UserClient   pb.UserServiceClient
    // DataClient   pb.DataServiceClient
    // Logger       *zap.Logger
    // Cache        *redis.Client
}
```

**Avantages:**

1. **Performance**: Une seule connexion partagée
2. **Testabilité**: Facile d'injecter des mocks pour les tests
3. **Maintenabilité**: Toutes les dépendances sont explicites

**Utilisation dans les resolvers:**

```go
// services/api-gateway/graph/resolvers_impl.go

func (r *Resolver) CreateDeviceImpl(
    ctx context.Context,
    input model.CreateDeviceInput,
) (*model.Device, error) {
    // r.DeviceClient est déjà initialisé
    // Pas besoin de créer une connexion!

    req := &pb.CreateDeviceRequest{
        Name: input.Name,
        Type: input.Type,
    }

    resp, err := r.DeviceClient.CreateDevice(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to create device: %w", err)
    }

    return protoToGraphQLDevice(resp.Device), nil
}
```

#### 8.5 Conversion de types: le challenge de l'intégration

**Le problème:**

Les types Protobuf et GraphQL ne sont pas identiques. Par exemple:

```protobuf
// Protobuf utilise des maps
message Device {
    map<string, string> metadata = 7;
}
```

```graphql
# GraphQL utilise des listes de paires clé-valeur
type Device {
    metadata: [MetadataEntry!]!
}

type MetadataEntry {
    key: String!
    value: String!
}
```

Pourquoi cette différence?
- **Protobuf**: Maps pour l'efficacité (accès O(1))
- **GraphQL**: Listes pour la compatibilité avec JSON et la facilité d'utilisation dans les frontends

**Solution: Fonctions de conversion**

```go
// services/api-gateway/graph/resolvers_impl.go

// Protobuf → GraphQL
func protoToGraphQLDevice(d *pb.Device) *model.Device {
    if d == nil {
        return nil
    }

    // Convertir map[string]string → []*MetadataEntry
    metadata := make([]*model.MetadataEntry, 0, len(d.Metadata))
    for key, value := range d.Metadata {
        metadata = append(metadata, &model.MetadataEntry{
            Key:   key,
            Value: value,
        })
    }

    return &model.Device{
        ID:        d.Id,
        Name:      d.Name,
        Type:      d.Type,
        Status:    protoToGraphQLStatus(d.Status),
        CreatedAt: int(d.CreatedAt),
        LastSeen:  int(d.LastSeen),
        Metadata:  metadata,
    }
}

// GraphQL → Protobuf
func graphQLToProtoMetadata(input []*model.MetadataEntryInput) map[string]string {
    metadata := make(map[string]string)
    if input != nil {
        for _, entry := range input {
            if entry != nil {
                metadata[entry.Key] = entry.Value
            }
        }
    }
    return metadata
}

// Conversion d'enum
func protoToGraphQLStatus(status pb.DeviceStatus) model.DeviceStatus {
    switch status {
    case pb.DeviceStatus_ONLINE:
        return model.DeviceStatusOnline
    case pb.DeviceStatus_OFFLINE:
        return model.DeviceStatusOffline
    case pb.DeviceStatus_ERROR:
        return model.DeviceStatusError
    case pb.DeviceStatus_MAINTENANCE:
        return model.DeviceStatusMaintenance
    default:
        return model.DeviceStatusUnknown
    }
}
```

**Patron de conception:**

Ces fonctions suivent le patron **Adapter** (Adaptateur):
- On a deux interfaces incompatibles (Protobuf et GraphQL)
- On crée un adaptateur qui traduit de l'une à l'autre
- Le reste du code n'a pas besoin de connaître les détails de la conversion

#### 8.6 Exemple complet: implémenter CreateDevice

Maintenant, assemblons tout pour voir comment une mutation complète fonctionne:

```go
// services/api-gateway/graph/resolvers_impl.go

package graph

import (
    "context"
    "fmt"

    pb "github.com/yourusername/iot-platform/shared/proto"
    "github.com/yourusername/iot-platform/services/api-gateway/graph/model"
)

// CreateDeviceImpl implémente la logique de création d'un device
func (r *Resolver) CreateDeviceImpl(
    ctx context.Context,
    input model.CreateDeviceInput,
) (*model.Device, error) {

    // ─────────────────────────────────────────────────────────
    // ÉTAPE 1: Validation (optionnelle, souvent le service le fait)
    // ─────────────────────────────────────────────────────────
    if input.Name == "" {
        return nil, fmt.Errorf("device name is required")
    }

    // ─────────────────────────────────────────────────────────
    // ÉTAPE 2: Conversion GraphQL Input → Protobuf Request
    // ─────────────────────────────────────────────────────────
    req := &pb.CreateDeviceRequest{
        Name:     input.Name,
        Type:     input.Type,
        Metadata: graphQLToProtoMetadata(input.Metadata),
    }

    // ─────────────────────────────────────────────────────────
    // ÉTAPE 3: Appel gRPC au Device Manager
    // ─────────────────────────────────────────────────────────
    resp, err := r.DeviceClient.CreateDevice(ctx, req)
    if err != nil {
        // Gestion d'erreur: on pourrait analyser le code d'erreur gRPC
        // et retourner une erreur GraphQL appropriée
        return nil, fmt.Errorf("failed to create device: %w", err)
    }

    // ─────────────────────────────────────────────────────────
    // ÉTAPE 4: Conversion Protobuf Response → GraphQL Type
    // ─────────────────────────────────────────────────────────
    device := protoToGraphQLDevice(resp.Device)

    // ─────────────────────────────────────────────────────────
    // ÉTAPE 5: Retour du résultat
    // ─────────────────────────────────────────────────────────
    return device, nil
}
```

**Le flux complet visualisé:**

```
Client GraphQL
│
│ mutation {
│   createDevice(input: {
│     name: "Sensor"
│     type: "temp"
│   }) { id name }
│ }
│
▼
┌─────────────────────────────────────────────────────────┐
│ GraphQL Server (généré)                                 │
│ • Parse la requête                                      │
│ • Valide contre le schéma                              │
│ • Appelle le resolver                                  │
└───────────────────┬─────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────┐
│ schema.resolvers.go (stub généré)                       │
│                                                         │
│ func (r *mutationResolver) CreateDevice(...) {          │
│     return r.Resolver.CreateDeviceImpl(ctx, input)      │
│ }                                                       │
└───────────────────┬─────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────┐
│ resolvers_impl.go (notre code)                          │
│                                                         │
│ func (r *Resolver) CreateDeviceImpl(...) {              │
│     req := &pb.CreateDeviceRequest{...}                 │
│     resp, err := r.DeviceClient.CreateDevice(ctx, req)  │
│     return protoToGraphQLDevice(resp.Device), nil       │
│ }                                                       │
└───────────────────┬─────────────────────────────────────┘
                    │
                    │ gRPC call (HTTP/2 + Protobuf)
                    ▼
┌─────────────────────────────────────────────────────────┐
│ Device Manager (service gRPC)                           │
│ • Valide                                               │
│ • Crée le device                                       │
│ • Stocke en DB                                         │
│ • Retourne le résultat                                 │
└─────────────────────────────────────────────────────────┘
```

#### 8.7 Initialisation dans main.go

Le fichier `main.go` est le chef d'orchestre qui initialise tout:

```go
// services/api-gateway/main.go

package main

import (
    "log"
    "net/http"
    "os"

    "github.com/99designs/gqlgen/graphql/handler"
    "github.com/99designs/gqlgen/graphql/playground"

    "github.com/yourusername/iot-platform/services/api-gateway/graph"
    "github.com/yourusername/iot-platform/services/api-gateway/graph/generated"
    grpcClient "github.com/yourusername/iot-platform/services/api-gateway/grpc"
)

func main() {
    // ─────────────────────────────────────────────────────────
    // 1. Configuration depuis les variables d'environnement
    // ─────────────────────────────────────────────────────────
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"  // Valeur par défaut
    }

    deviceManagerAddr := os.Getenv("DEVICE_MANAGER_ADDR")
    if deviceManagerAddr == "" {
        deviceManagerAddr = "localhost:8081"
    }

    log.Printf("🚀 Starting API Gateway...")
    log.Printf("   Port: %s", port)
    log.Printf("   Device Manager: %s", deviceManagerAddr)

    // ─────────────────────────────────────────────────────────
    // 2. Connexion au Device Manager (gRPC)
    // ─────────────────────────────────────────────────────────
    deviceClient, err := grpcClient.NewDeviceClient(deviceManagerAddr)
    if err != nil {
        log.Fatalf("❌ Failed to connect to Device Manager: %v", err)
    }
    defer deviceClient.Close()  // Fermeture propre au shutdown

    // ─────────────────────────────────────────────────────────
    // 3. Création du serveur GraphQL avec injection des dépendances
    // ─────────────────────────────────────────────────────────
    srv := handler.NewDefaultServer(
        generated.NewExecutableSchema(
            generated.Config{
                Resolvers: &graph.Resolver{
                    DeviceClient: deviceClient.GetClient(),
                },
            },
        ),
    )

    // ─────────────────────────────────────────────────────────
    // 4. Configuration des routes HTTP
    // ─────────────────────────────────────────────────────────

    // GraphQL Playground (interface web pour tester)
    http.Handle("/", playground.Handler("GraphQL Playground", "/query"))

    // Endpoint GraphQL (où les clients envoient leurs requêtes)
    http.Handle("/query", srv)

    // Health check (pour Kubernetes, load balancers, etc.)
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })

    // ─────────────────────────────────────────────────────────
    // 5. Démarrage du serveur HTTP
    // ─────────────────────────────────────────────────────────
    log.Printf("✅ GraphQL Playground: http://localhost:%s", port)
    log.Printf("✅ GraphQL Endpoint:   http://localhost:%s/query", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}
```

**Points clés:**

1. **Configuration par environnement**: Permet de changer les ports/adresses sans recompiler
2. **Fail-fast**: Si la connexion au Device Manager échoue, on arrête tout immédiatement
3. **Graceful shutdown**: `defer deviceClient.Close()` garantit la fermeture propre
4. **Health check**: Indispensable pour les orchestrateurs (Kubernetes, Docker Swarm)

---

### 9. Synchronisation et sécurité des données

#### 9.1 Le problème du partage de données

Imaginez une situation du quotidien: plusieurs personnes veulent modifier le même document Google Docs en même temps. Si le système n'est pas conçu pour gérer cela, on obtient un chaos de modifications qui s'écrasent mutuellement.

En programmation, c'est exactement le même problème. Quand plusieurs requêtes HTTP arrivent simultanément sur notre serveur Go, elles sont traitées en **parallèle** (goroutines). Si toutes modifient la même structure de données (ex: une map), on a un problème.

**Exemple concret:**

```go
// ❌ CODE DANGEREUX
type DeviceServer struct {
    devices map[string]*pb.Device
}

func (s *DeviceServer) CreateDevice(...) {
    device := &pb.Device{Id: "123", ...}
    s.devices[device.Id] = device  // RACE CONDITION!
}

func (s *DeviceServer) GetDevice(...) {
    device := s.devices[id]  // RACE CONDITION!
    return device
}
```

**Que se passe-t-il?**

```
Temps  Goroutine 1              Goroutine 2              État de la map
─────  ─────────────────────   ─────────────────────   ────────────────
t0     CreateDevice("123")     -                       {}
t1     Lit la map              GetDevice("123")        {}
t2     Écrit "123"             Lit la map              {123: ...}
t3     -                       Lit "123"               {123: ...}

Scenario normal ✅

Temps  Goroutine 1              Goroutine 2              État de la map
─────  ─────────────────────   ─────────────────────   ────────────────
t0     CreateDevice("123")     CreateDevice("456")     {}
t1     Lit la map              Lit la map              {}
t2     Écrit "123"             Écrit "456"             💥 CRASH!
t3     -                       -                       Map corrompue

Race condition ❌
```

Go détecte ces situations avec le **race detector**:

```bash
go run -race main.go

==================
WARNING: DATA RACE
Write at 0x... by goroutine 1:
Read at 0x... by goroutine 2:
==================
```

#### 9.2 La solution: sync.RWMutex

Un **Mutex** (Mutual Exclusion) est comme un verrou. Un **RWMutex** (Read-Write Mutex) est un verrou intelligent qui comprend la différence entre lire et écrire.

**Analogie:**

Imaginez une bibliothèque:
- **Lecture**: Plusieurs personnes peuvent lire des livres en même temps (pas de problème)
- **Écriture**: Si quelqu'un modifie le catalogue, personne d'autre ne doit lire ou modifier (exclusivité)

```go
// ✅ CODE SÛR
type DeviceServer struct {
    mu      sync.RWMutex              // Le verrou
    devices map[string]*pb.Device     // La donnée protégée
}

// Écriture: Lock exclusif
func (s *DeviceServer) CreateDevice(
    ctx context.Context,
    req *pb.CreateDeviceRequest,
) (*pb.CreateDeviceResponse, error) {
    device := &pb.Device{
        Id:   uuid.New().String(),
        Name: req.Name,
        // ...
    }

    s.mu.Lock()                     // 🔒 Acquiert le verrou (exclusif)
    s.devices[device.Id] = device   // Modification protégée
    s.mu.Unlock()                   // 🔓 Libère le verrou

    return &pb.CreateDeviceResponse{Device: device}, nil
}

// Lecture: RLock partagé
func (s *DeviceServer) GetDevice(
    ctx context.Context,
    req *pb.GetDeviceRequest,
) (*pb.GetDeviceResponse, error) {
    s.mu.RLock()                    // 🔒 Acquiert le verrou (partagé)
    device := s.devices[req.Id]     // Lecture protégée
    s.mu.RUnlock()                  // 🔓 Libère le verrou

    if device == nil {
        return nil, status.Error(codes.NotFound, "device not found")
    }

    return &pb.GetDeviceResponse{Device: device}, nil
}
```

#### 9.3 Lock vs RLock: comprendre la différence

| Opération | Type de verrou | Bloque Lock? | Bloque RLock? | Usage |
|-----------|---------------|--------------|---------------|-------|
| `Lock()` | Exclusif | ✅ Oui | ✅ Oui | Écriture (Create, Update, Delete) |
| `RLock()` | Partagé | ✅ Oui | ❌ Non | Lecture (Get, List) |

**Scénario avec plusieurs goroutines:**

```
Temps  Goroutine 1          Goroutine 2          Goroutine 3          État du verrou
─────  ──────────────────   ──────────────────   ──────────────────   ──────────────
t0     RLock() ✅           -                    -                    Read (1)
t1     Lit device           RLock() ✅           -                    Read (2)
t2     Lit device           Lit device           RLock() ✅           Read (3)
t3     RUnlock()            Lit device           Lit device           Read (2)
t4     -                    RUnlock()            Lit device           Read (1)
t5     -                    -                    RUnlock()            Unlocked
t6     -                    -                    Lock() ✅            Write (1)
t7     RLock() ⏳ attend    -                    Écrit device         Write (1)
t8     ⏳ attend            -                    Unlock()             Unlocked
t9     RLock() ✅           -                    -                    Read (1)
```

**Points clés:**
- Plusieurs `RLock()` peuvent coexister
- Un `Lock()` attend que tous les `RLock()` et autres `Lock()` se terminent
- Quand quelqu'un a un `Lock()`, personne d'autre ne peut acquérir quoi que ce soit

#### 9.4 Bonnes pratiques avec les Mutex

**Pratique 1: Utiliser `defer` pour garantir le Unlock**

```go
// ❌ Dangereux
func (s *DeviceServer) GetDevice(...) {
    s.mu.RLock()
    device := s.devices[req.Id]

    if device == nil {
        return nil, errors.New("not found")  // 💥 RUnlock jamais appelé!
    }

    s.mu.RUnlock()
    return device, nil
}

// ✅ Sûr
func (s *DeviceServer) GetDevice(...) {
    s.mu.RLock()
    defer s.mu.RUnlock()  // Garantit Unlock même en cas d'erreur

    device := s.devices[req.Id]
    if device == nil {
        return nil, errors.New("not found")  // RUnlock sera appelé
    }

    return device, nil
}
```

**Pratique 2: Scope minimal du Lock**

```go
// ❌ Lock trop long
func (s *DeviceServer) CreateDevice(...) {
    s.mu.Lock()
    defer s.mu.Unlock()

    // Génération d'UUID (lent)
    id := uuid.New().String()

    // Validation (peut être lente)
    if err := s.validator.Validate(req); err != nil {
        return nil, err
    }

    // Appel API externe (très lent!)
    metadata, _ := s.externalAPI.FetchMetadata(id)

    // Enfin, l'écriture
    s.devices[id] = &pb.Device{...}
}

// ✅ Lock minimal
func (s *DeviceServer) CreateDevice(...) {
    // Tout le travail AVANT le lock
    id := uuid.New().String()

    if err := s.validator.Validate(req); err != nil {
        return nil, err
    }

    metadata, _ := s.externalAPI.FetchMetadata(id)

    device := &pb.Device{Id: id, Metadata: metadata, ...}

    // Lock uniquement pour l'écriture
    s.mu.Lock()
    s.devices[id] = device
    s.mu.Unlock()

    return device, nil
}
```

**Pourquoi?** Plus le Lock dure longtemps, plus les autres goroutines attendent. On veut minimiser le temps passé sous Lock.

**Pratique 3: Éviter les deadlocks**

```go
// ❌ DEADLOCK GARANTI
func (s *DeviceServer) BadMethod() {
    s.mu.Lock()
    defer s.mu.Unlock()

    // Cette méthode essaie aussi de Lock!
    s.AnotherMethod()  // 💥 Attend indéfiniment
}

func (s *DeviceServer) AnotherMethod() {
    s.mu.Lock()  // Impossible, déjà locked!
    defer s.mu.Unlock()
    // ...
}

// ✅ Séparation des responsabilités
func (s *DeviceServer) GoodMethod() {
    s.mu.Lock()
    // Faire le travail directement ici
    s.mu.Unlock()
}

// Ou: méthode interne sans Lock
func (s *DeviceServer) PublicMethod() {
    s.mu.Lock()
    defer s.mu.Unlock()

    s.internalMethodNoLock()  // ✅ OK
}

func (s *DeviceServer) internalMethodNoLock() {
    // Assume que le caller a déjà le lock
    // Convention: suffixe "NoLock" ou "Unsafe" dans le nom
}
```

#### 9.5 Alternatives et quand les utiliser

**Alternative 1: Channels**

Pour la communication entre goroutines, souvent mieux que Mutex:

```go
// Pattern: Une goroutine "propriétaire" de la donnée
type DeviceManager struct {
    requests chan DeviceRequest
}

func (m *DeviceManager) Run() {
    devices := make(map[string]*Device)  // Pas besoin de Mutex!

    for req := range m.requests {
        switch req.Type {
        case "create":
            devices[req.ID] = req.Device
        case "get":
            req.Response <- devices[req.ID]
        }
    }
}
```

**Alternative 2: sync.Map**

Map thread-safe intégrée (mais moins performante que RWMutex pour la plupart des cas):

```go
type DeviceServer struct {
    devices sync.Map  // Map thread-safe
}

func (s *DeviceServer) CreateDevice(...) {
    device := &pb.Device{...}
    s.devices.Store(device.Id, device)  // Thread-safe
}

func (s *DeviceServer) GetDevice(...) {
    value, ok := s.devices.Load(req.Id)  // Thread-safe
    if !ok {
        return nil, errors.New("not found")
    }
    device := value.(*pb.Device)
    return device, nil
}
```

**Quand utiliser sync.Map?**
- Quand les clés sont écrites une fois et lues souvent
- Quand différentes goroutines lisent/écrivent des clés différentes
- Sinon, `RWMutex` est généralement plus rapide

---

## Partie IV - Déploiement et bonnes pratiques

### 10. Containerisation avec Docker

#### 10.1 Pourquoi Docker?

**Le problème classique: "Ça marche sur ma machine"**

```
Développeur: "Mon code marche parfaitement!"
Ops:         "Chez moi ça crash, quelle version de Go tu utilises?"
Développeur: "Go 1.23 avec PostgreSQL 15"
Ops:         "J'ai Go 1.20 et PostgreSQL 14..."
```

**Docker résout ce problème** en créant un **conteneur**: un environnement isolé qui contient:
- Votre application
- Toutes ses dépendances
- La version exacte de chaque outil
- Les variables d'environnement

**Analogie:**
Un conteneur Docker est comme un conteneur de transport maritime:
- Standardisé (même taille, même interface)
- Peut être déplacé partout (laptop, serveur, cloud)
- Contient tout ce dont vous avez besoin
- Isolé du reste (ce qui se passe dedans ne sort pas)

#### 10.2 Docker Compose: orchestrer plusieurs conteneurs

Notre plateforme IoT a besoin de plusieurs services (API Gateway, Device Manager, PostgreSQL, Redis). Docker Compose permet de les lancer tous ensemble avec une seule commande.

```yaml
# docker-compose.yml

version: '3.9'

services:
  # ────────────────────────────────────────────────────────
  # Base de données relationnelle (TimescaleDB = PostgreSQL optimisé pour séries temporelles)
  # ────────────────────────────────────────────────────────
  postgres:
    image: timescale/timescaledb:latest-pg16
    container_name: iot-postgres
    ports:
      - "5432:5432"  # Port exposé: host:container
    environment:
      POSTGRES_USER: iot_user
      POSTGRES_PASSWORD: iot_password
      POSTGRES_DB: iot_platform
    volumes:
      # Volume nommé: les données persistent même si on supprime le conteneur
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      # Vérification que la DB est prête
      test: ["CMD-SHELL", "pg_isready -U iot_user"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - iot-network

  # ────────────────────────────────────────────────────────
  # Cache en mémoire
  # ────────────────────────────────────────────────────────
  redis:
    image: redis:7-alpine
    container_name: iot-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes  # Persistance sur disque
    networks:
      - iot-network

  # ────────────────────────────────────────────────────────
  # Message broker MQTT pour les devices IoT
  # ────────────────────────────────────────────────────────
  mosquitto:
    image: eclipse-mosquitto:2
    container_name: iot-mosquitto
    ports:
      - "1883:1883"   # MQTT
      - "9001:9001"   # WebSocket
    volumes:
      - ./infrastructure/docker/mosquitto/mosquitto.conf:/mosquitto/config/mosquitto.conf
      - mosquitto_data:/mosquitto/data
      - mosquitto_logs:/mosquitto/log
    networks:
      - iot-network

# ────────────────────────────────────────────────────────
# Volumes nommés (persistance des données)
# ────────────────────────────────────────────────────────
volumes:
  postgres_data:
    name: iot_postgres_data
  redis_data:
    name: iot_redis_data
  mosquitto_data:
    name: iot_mosquitto_data
  mosquitto_logs:
    name: iot_mosquitto_logs

# ────────────────────────────────────────────────────────
# Réseau pour que les conteneurs communiquent
# ────────────────────────────────────────────────────────
networks:
  iot-network:
    name: iot-platform-network
    driver: bridge
```

**Explications des concepts:**

**1. Images vs Conteneurs**

```
Image:      Modèle (template) immuable
            Comme un ISO de CD d'installation

Conteneur:  Instance en cours d'exécution
            Comme un programme lancé depuis le CD
```

**2. Ports**

```yaml
ports:
  - "5432:5432"
     ↑     ↑
  host   container
```

- Le service écoute sur le port 5432 DANS le conteneur
- On le rend accessible sur le port 5432 de la machine hôte
- Vous pouvez changer: `"5433:5432"` pour éviter les conflits

**3. Volumes**

```yaml
volumes:
  - postgres_data:/var/lib/postgresql/data
       ↑                    ↑
    volume nommé      path dans conteneur
```

**Sans volume:**
```
docker-compose up    → Données créées
docker-compose down  → Données PERDUES
```

**Avec volume:**
```
docker-compose up    → Données créées dans volume
docker-compose down  → Volume préservé
docker-compose up    → Données toujours là!
```

**4. Networks**

Les conteneurs sur le même réseau peuvent communiquer entre eux par leur nom:

```go
// Dans Device Manager, on peut utiliser:
db, err := sql.Open("postgres", "host=postgres user=iot_user ...")
//                                      ↑
//                               Nom du service Docker
```

Docker fait automatiquement la résolution DNS: `postgres` → IP du conteneur.

#### 10.3 Commandes Docker essentielles

```bash
# ════════════════════════════════════════════════════════
# Démarrage et arrêt
# ════════════════════════════════════════════════════════

# Démarrer tous les services (en arrière-plan)
docker-compose up -d

# Voir les logs en temps réel
docker-compose logs -f

# Logs d'un service spécifique
docker-compose logs -f postgres

# Arrêter tous les services (garde les volumes)
docker-compose down

# Arrêter ET supprimer les volumes (⚠️ perte de données!)
docker-compose down -v

# ════════════════════════════════════════════════════════
# Inspection
# ════════════════════════════════════════════════════════

# Voir les conteneurs en cours d'exécution
docker-compose ps

# Détails d'un conteneur
docker inspect iot-postgres

# Utilisation des ressources (CPU, RAM)
docker stats

# ════════════════════════════════════════════════════════
# Exécution de commandes dans un conteneur
# ════════════════════════════════════════════════════════

# Se connecter à PostgreSQL
docker exec -it iot-postgres psql -U iot_user -d iot_platform

# Shell dans le conteneur
docker exec -it iot-postgres bash

# Commande ponctuelle
docker exec iot-postgres pg_dump -U iot_user iot_platform > backup.sql

# ════════════════════════════════════════════════════════
# Nettoyage
# ════════════════════════════════════════════════════════

# Supprimer les conteneurs arrêtés
docker-compose rm

# Supprimer les images non utilisées
docker image prune

# Nettoyage complet (⚠️ supprime tout ce qui est inutilisé)
docker system prune -a --volumes
```

---

### 11. Patterns de conception et bonnes pratiques

#### 11.1 Error handling en Go

Go a une philosophie différente de la gestion d'erreurs par rapport à d'autres langages (pas d'exceptions).

**Principe: Les erreurs sont des valeurs**

```go
// ❌ Langages avec exceptions (Java, Python)
try {
    device = createDevice(name)
    saveToDatabase(device)
    sendNotification(device)
} catch (Exception e) {
    // On ne sait pas où l'erreur s'est produite
}

// ✅ Go: explicite
device, err := createDevice(name)
if err != nil {
    return fmt.Errorf("failed to create device: %w", err)
}

err = saveToDatabase(device)
if err != nil {
    return fmt.Errorf("failed to save device: %w", err)
}

err = sendNotification(device)
if err != nil {
    // C'est pas grave si la notification échoue, on continue
    log.Printf("warning: failed to send notification: %v", err)
}
```

**Avantages:**
- Flux d'exécution clair (pas de sauts invisibles)
- On est forcé de gérer chaque erreur
- Difficile d'oublier de gérer une erreur

**Pattern: Wrapping d'erreurs**

```go
func (s *DeviceServer) CreateDevice(...) (*Device, error) {
    // Étape 1
    if err := s.validate(req); err != nil {
        // Ajouter du contexte à l'erreur
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    // Étape 2
    device, err := s.repository.Save(device)
    if err != nil {
        // Wrap avec contexte
        return nil, fmt.Errorf("failed to save to database: %w", err)
    }

    return device, nil
}

// Si erreur, on obtient une stack d'erreurs:
// "failed to save to database: connection refused: dial tcp 127.0.0.1:5432: connect: connection refused"
//  ↑                           ↑                     ↑
//  Contexte niveau service    Contexte repo         Erreur originale
```

**Pattern: Custom errors**

```go
// Définir des erreurs spécifiques
var (
    ErrDeviceNotFound     = errors.New("device not found")
    ErrInvalidDeviceType  = errors.New("invalid device type")
    ErrDuplicateDevice    = errors.New("device already exists")
)

func (s *DeviceServer) GetDevice(...) (*Device, error) {
    device := s.devices[id]
    if device == nil {
        // Utiliser l'erreur prédéfinie
        return nil, ErrDeviceNotFound
    }
    return device, nil
}

// Le caller peut tester l'erreur spécifique
device, err := server.GetDevice(id)
if errors.Is(err, ErrDeviceNotFound) {
    // Gérer spécifiquement le cas "not found"
    return status.Error(codes.NotFound, "device not found")
}
```

#### 11.2 Codes d'erreur gRPC

gRPC standardise les codes d'erreur (similaire à HTTP):

```go
import "google.golang.org/grpc/codes"
import "google.golang.org/grpc/status"

func (s *DeviceServer) CreateDevice(...) (*Device, error) {
    // Validation
    if req.Name == "" {
        return nil, status.Error(codes.InvalidArgument, "name is required")
    }

    // Vérifier existence
    if s.deviceExists(req.Name) {
        return nil, status.Error(codes.AlreadyExists, "device name already exists")
    }

    // Erreur de base de données
    device, err := s.repository.Save(...)
    if err != nil {
        log.Printf("Database error: %v", err)
        return nil, status.Error(codes.Internal, "internal server error")
    }

    return device, nil
}
```

**Mapping codes gRPC ↔ HTTP:**

| Code gRPC | Description | HTTP équivalent |
|-----------|-------------|-----------------|
| `OK` | Succès | 200 OK |
| `InvalidArgument` | Arguments invalides | 400 Bad Request |
| `NotFound` | Ressource introuvable | 404 Not Found |
| `AlreadyExists` | Duplication | 409 Conflict |
| `PermissionDenied` | Pas de permissions | 403 Forbidden |
| `Unauthenticated` | Authentification requise | 401 Unauthorized |
| `Internal` | Erreur serveur | 500 Internal Server Error |
| `Unavailable` | Service indisponible | 503 Service Unavailable |

#### 11.3 Context: le fil rouge des requêtes

Le `context.Context` en Go est un concept fondamental mais souvent mal compris.

**À quoi ça sert?**

1. **Propagation d'annulation**
2. **Timeout**
3. **Métadonnées** (request ID, user ID, etc.)

**Exemple: Timeout**

```go
// Sans context: risque de bloquer indéfiniment
func fetchDataFromExternalAPI() (*Data, error) {
    resp, err := http.Get("https://slow-api.com/data")  // Peut prendre 5 minutes!
    // ...
}

// Avec context: timeout automatique
func fetchDataFromExternalAPI(ctx context.Context) (*Data, error) {
    req, _ := http.NewRequestWithContext(ctx, "GET", "https://slow-api.com/data", nil)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Si timeout dépassé, err sera context.DeadlineExceeded
        return nil, err
    }
    // ...
}

// Utilisation avec timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

data, err := fetchDataFromExternalAPI(ctx)
if errors.Is(err, context.DeadlineExceeded) {
    log.Println("API trop lente, on abandonne")
}
```

**Exemple: Propagation d'annulation**

```go
func (s *DeviceServer) ProcessDevices(ctx context.Context) error {
    devices, err := s.ListDevices(ctx)
    if err != nil {
        return err
    }

    for _, device := range devices {
        // Vérifier si le client a annulé
        select {
        case <-ctx.Done():
            return ctx.Err()  // context.Canceled
        default:
        }

        // Traiter le device (propager le context)
        if err := s.ProcessDevice(ctx, device); err != nil {
            return err
        }
    }

    return nil
}
```

**Règles d'or du Context:**

1. **Toujours premier paramètre**: `func Foo(ctx context.Context, ...)`
2. **Ne jamais stocker dans une struct**: Passer en paramètre
3. **Propager partout**: Chaque appel doit recevoir le context
4. **Ne jamais passer `nil`**: Utiliser `context.Background()` si pas de context parent

---

### 12. Outils de développement et workflow

#### 12.1 Organisation du Makefile

Notre Makefile est organisé en sections logiques pour faciliter le développement.

**Changement récent:** Nous avons amélioré la commande `make help` pour afficher les commandes organisées par sections, et simplifié `make dev` pour lancer les deux services ensemble.

```makefile
# Makefile

# ════════════════════════════════════════════════════════
# Configuration
# ════════════════════════════════════════════════════════

.PHONY: help
.DEFAULT_GOAL := help

# Couleurs pour l'affichage
CYAN := \033[36m
GREEN := \033[32m
YELLOW := \033[33m
RESET := \033[0m

# ════════════════════════════════════════════════════════
# Help - Affichage organisé par sections
# ════════════════════════════════════════════════════════

help:
	@echo ""
	@echo "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(GREEN)  IoT Platform - Commandes disponibles$(RESET)"
	@echo "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""
	@echo "$(YELLOW)🚀 Développement$(RESET)"
	@echo "  $(CYAN)dev$(RESET)                  Lance API Gateway + Device Manager en parallèle"
	@echo "  $(CYAN)api-gateway$(RESET)          Lance uniquement l'API Gateway"
	@echo "  $(CYAN)device-manager$(RESET)       Lance uniquement le Device Manager"
	@echo ""
	@echo "$(YELLOW)🔧 Installation$(RESET)"
	@echo "  $(CYAN)setup$(RESET)                Installe les outils et dépendances"
	@echo "  $(CYAN)deps$(RESET)                 Télécharge les dépendances Go"
	@echo ""
	@echo "$(YELLOW)⚙️  Génération de code$(RESET)"
	@echo "  $(CYAN)generate$(RESET)             Génère tout le code (proto + GraphQL)"
	@echo "  $(CYAN)generate-proto$(RESET)       Génère uniquement le code Protocol Buffers"
	@echo "  $(CYAN)generate-graphql$(RESET)     Génère uniquement le code GraphQL"
	@echo ""
	@echo "$(YELLOW)🐳 Docker$(RESET)"
	@echo "  $(CYAN)start$(RESET)                Démarre l'infrastructure Docker"
	@echo "  $(CYAN)stop$(RESET)                 Arrête l'infrastructure Docker"
	@echo "  $(CYAN)status$(RESET)               Affiche le statut des conteneurs"
	@echo "  $(CYAN)logs$(RESET)                 Affiche les logs Docker"
	@echo ""
	@echo "$(YELLOW)🧹 Nettoyage$(RESET)"
	@echo "  $(CYAN)clean$(RESET)                Nettoie les fichiers de build"
	@echo "  $(CYAN)clean-all$(RESET)            Nettoie tout (build + volumes Docker)"
	@echo ""
	@echo "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""

# ════════════════════════════════════════════════════════
# Développement
# ════════════════════════════════════════════════════════

# NOUVEAU: Commande simplifiée pour lancer les deux services
dev:
	@echo "$(GREEN)🚀 Démarrage API Gateway + Device Manager...$(RESET)"
	@trap 'kill 0' INT; \
	cd services/device-manager && go run main.go & \
	cd services/api-gateway && go run main.go & \
	wait

api-gateway:
	@echo "$(GREEN)🚀 Démarrage API Gateway...$(RESET)"
	cd services/api-gateway && go run main.go

device-manager:
	@echo "$(GREEN)🚀 Démarrage Device Manager...$(RESET)"
	cd services/device-manager && go run main.go

# ════════════════════════════════════════════════════════
# Installation
# ════════════════════════════════════════════════════════

setup:
	@echo "$(GREEN)📦 Installation des outils...$(RESET)"
	@# Protoc
	@which protoc > /dev/null || (echo "Installer protoc: https://grpc.io/docs/protoc-installation/" && exit 1)
	@# Go plugins pour protoc
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@# gqlgen
	go install github.com/99designs/gqlgen@latest
	@echo "$(GREEN)✅ Installation terminée!$(RESET)"

deps:
	@echo "$(GREEN)📦 Téléchargement des dépendances...$(RESET)"
	cd shared/proto && go mod download
	cd services/device-manager && go mod download
	cd services/api-gateway && go mod download

# ════════════════════════════════════════════════════════
# Génération de code
# ════════════════════════════════════════════════════════

generate: generate-proto generate-graphql

generate-proto:
	@echo "$(GREEN)⚙️  Génération du code Protocol Buffers...$(RESET)"
	cd shared/proto && ./generate.sh

generate-graphql:
	@echo "$(GREEN)⚙️  Génération du code GraphQL...$(RESET)"
	cd services/api-gateway && gqlgen generate

# ════════════════════════════════════════════════════════
# Docker
# ════════════════════════════════════════════════════════

start:
	@echo "$(GREEN)🐳 Démarrage de l'infrastructure...$(RESET)"
	docker-compose up -d

stop:
	@echo "$(YELLOW)🛑 Arrêt de l'infrastructure...$(RESET)"
	docker-compose down

status:
	@echo "$(CYAN)📊 Statut des conteneurs:$(RESET)"
	@docker-compose ps

logs:
	docker-compose logs -f

# ════════════════════════════════════════════════════════
# Nettoyage
# ════════════════════════════════════════════════════════

clean:
	@echo "$(YELLOW)🧹 Nettoyage des fichiers de build...$(RESET)"
	find . -name "*.exe" -type f -delete
	find . -name "api-gateway" -type f -delete
	find . -name "device-manager" -type f -delete

clean-all: clean stop
	@echo "$(YELLOW)🧹 Nettoyage complet (+ volumes Docker)...$(RESET)"
	docker-compose down -v
```

**Points clés des améliorations:**

1. **`make help` organisé**: Les commandes sont regroupées par catégories (Développement, Installation, etc.)
2. **`make dev` simplifié**: Lance les deux services en parallèle avec un seul Ctrl+C pour tout arrêter
3. **Couleurs**: Meilleure lisibilité dans le terminal
4. **Descriptions claires**: Chaque commande explique ce qu'elle fait

#### 12.2 Workflow de développement quotidien

**Jour 1: Démarrage d'un projet**

```bash
# 1. Cloner le repo
git clone <repo>
cd iot-platform

# 2. Installer les outils
make setup

# 3. Télécharger les dépendances
make deps

# 4. Démarrer l'infrastructure (bases de données, etc.)
make start

# 5. Attendre que tout soit prêt
make status

# 6. Lancer les services
make dev
```

**Workflow quotidien:**

```bash
# Matin: démarrer l'infrastructure
make start && make dev

# Modifier un schéma proto ou GraphQL
vim shared/proto/device.proto

# Régénérer le code
make generate

# Les services redémarrent automatiquement (si vous utilisez air ou similar)
# Sinon: Ctrl+C et `make dev` à nouveau

# Soir: tout éteindre
Ctrl+C (arrête les services)
make stop (arrête Docker)
```

**Workflow de feature:**

```bash
# 1. Créer une branche
git checkout -b feature/add-device-location

# 2. Modifier le schéma
vim shared/proto/device.proto
# Ajouter: string location = 8;

# 3. Régénérer
make generate

# 4. Implémenter (le compilateur vous dira ce qui manque)
vim services/device-manager/server.go
vim services/api-gateway/graph/schema.graphql
vim services/api-gateway/graph/resolvers_impl.go

# 5. Tester
make dev
# Tester manuellement dans GraphQL Playground

# 6. Commit
git add .
git commit -m "feat: Add device location field"

# 7. Push
git push origin feature/add-device-location
```

#### 12.3 Debugging et troubleshooting

**Problème 1: Service ne démarre pas**

```bash
# Vérifier les ports
lsof -i :8080
lsof -i :8081

# Tuer le process qui bloque
kill -9 <PID>
```

**Problème 2: Erreur de connexion gRPC**

```bash
# Vérifier que Device Manager tourne
curl http://localhost:8081  # Devrait retourner une erreur HTTP (normal, c'est gRPC)

# Tester la connexion gRPC
grpcurl -plaintext localhost:8081 list  # Liste les services
```

**Problème 3: Erreur de compilation après génération**

```bash
# Nettoyer et régénérer
rm -rf services/api-gateway/graph/generated
rm -rf shared/proto/*.pb.go
make generate

# Vérifier les dépendances
cd services/api-gateway && go mod tidy
```

---

## Prochaines étapes et ressources

### Ce que nous avons appris

Dans ce cours, nous avons couvert:

✅ **Fondements théoriques**
- Architecture microservices et quand l'utiliser
- Protocoles de communication (HTTP/1.1, HTTP/2)
- Comparaison REST vs GraphQL vs gRPC

✅ **Concepts clés**
- Protocol Buffers et définition d'interfaces
- GraphQL et flexibilité côté client
- Génération de code et typage fort

✅ **Mise en pratique**
- Architecture complète d'une plateforme IoT
- Implémentation de l'API Gateway
- Client gRPC persistant avec `grpc.NewClient` (pratique moderne)
- Synchronisation avec RWMutex

✅ **Déploiement**
- Containerisation avec Docker
- Orchestration avec Docker Compose
- Workflow de développement avec Make

### Ce qui reste à faire

**Court terme:**
- [ ] Remplacer le stockage en mémoire par PostgreSQL
- [ ] Ajouter l'authentification JWT
- [ ] Implémenter le rate limiting
- [ ] Ajouter des tests unitaires et d'intégration
- [ ] Logging structuré avec zerolog ou zap

**Moyen terme:**
- [ ] Service de collecte de données (Rust)
- [ ] Implémentation MQTT pour les devices IoT
- [ ] Dashboard web (React + Apollo Client)
- [ ] Métriques et monitoring (Prometheus + Grafana)
- [ ] Distributed tracing (Jaeger)

**Long terme:**
- [ ] Déploiement Kubernetes
- [ ] CI/CD avec GitHub Actions
- [ ] Auto-scaling basé sur les métriques
- [ ] Disaster recovery et backups automatiques

### Ressources pour aller plus loin

**Documentation officielle:**
- [gRPC Go](https://grpc.io/docs/languages/go/)
- [Protocol Buffers](https://protobuf.dev/)
- [GraphQL](https://graphql.org/learn/)
- [gqlgen](https://gqlgen.com/)
- [Docker](https://docs.docker.com/)

**Livres recommandés:**
- "Designing Data-Intensive Applications" de Martin Kleppmann
- "Building Microservices" de Sam Newman
- "The Go Programming Language" de Donovan & Kernighan

**Concepts avancés à explorer:**
1. **gRPC Streaming**: Bidirectionnel pour temps réel
2. **GraphQL Subscriptions**: WebSocket pour notifications live
3. **Event Sourcing**: Architecture événementielle
4. **CQRS**: Command Query Responsibility Segregation
5. **Service Mesh**: Istio, Linkerd pour observabilité

---

## Notes de version

**2026-01-09 - Version actuelle**
- ✅ Refonte complète avec approche pédagogique
- ✅ Migration `grpc.DialContext` → `grpc.NewClient`
- ✅ Amélioration du Makefile (help organisé, dev simplifié)
- ✅ Focus sur les concepts avant la technique
- ✅ Analogies et exemples concrets

**Prochaine mise à jour prévue:**
- Ajout d'une section sur PostgreSQL
- Diagrammes de séquence pour les flux de données
- Exemples de tests unitaires

---

**Note finale:** Ce document est vivant et évoluera avec le projet. N'hésitez pas à ajouter vos propres notes, questions et découvertes dans la section ci-dessous!

## Mes notes personnelles

_Espace réservé pour vos notes d'apprentissage..._
