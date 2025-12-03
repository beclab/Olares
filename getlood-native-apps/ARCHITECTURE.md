# Architecture Native Olares : Intégration AIOS & SmythOS via YAML

**Auteur** : Manus AI
**Date** : 17 novembre 2025
**Version** : 3.0.0

---

## 🎯 Vision Simplifiée

Au lieu de créer des couches MCP complexes, nous utilisons directement **l'infrastructure native d'Olares** :

- **Kubernetes d'Olares** pour l'orchestration
- **Docker d'Olares** pour les conteneurs
- **Réseau Olares** pour la communication
- **BFL Gateway** pour le routing

AIOS et SmythOS deviennent de **simples applications Olares rebrandées**, orchestrées uniquement par des **manifestes YAML**.

---

## 🏗️ Architecture Globale

```
┌──────────────────────────────────────────────────────────────┐
│  OLARES DESKTOP (Interface Utilisateur)                     │
│  - Files, Settings, Market, Desktop                         │
└────────────────────────┬─────────────────────────────────────┘
                         │
                         ↓ BFL Gateway (Routing HTTP/gRPC)
┌────────────────────────┼─────────────────────────────────────┐
│  APPLICATIONS OLARES (Rebrandées)                           │
│                        │                                     │
│  ┌──────────────────────────────────────────┐               │
│  │  Getlood Brain (AIOS Rebrandé)           │               │
│  │  - Namespace: getloodbrain-{user}        │               │
│  │  - URL: api.getloodbrain.{user}.local    │               │
│  │  - Services: Scheduler, Context, Memory  │               │
│  └──────────────────────────────────────────┘               │
│                                                              │
│  ┌──────────────────────────────────────────┐               │
│  │  Getlood Agents (SmythOS Rebrandé)       │               │
│  │  - Namespace: getloodagents-{user}       │               │
│  │  - URL: api.getloodagents.{user}.local   │               │
│  │  - Services: Agent Runtime, LLM Manager  │               │
│  └──────────────────────────────────────────┘               │
│                                                              │
│  ┌──────────────────────────────────────────┐               │
│  │  Getlood LLM (Ollama Rebrandé)           │               │
│  │  - Namespace: getloodllm-{user}          │               │
│  │  - URL: api.getloodllm.{user}.local      │               │
│  │  - Services: Ollama, vLLM                │               │
│  └──────────────────────────────────────────┘               │
└────────────────────────┼─────────────────────────────────────┘
                         │
                         ↓ Kubernetes Services
┌────────────────────────┼─────────────────────────────────────┐
│  OLARES KUBERNETES CLUSTER                                   │
│  - Namespaces par utilisateur                               │
│  - Services avec DNS interne                                │
│  - PersistentVolumes pour données                           │
└──────────────────────────────────────────────────────────────┘
```

---

## 🔑 Principes Clés

### 1. **Zéro Code Custom**

Tout est orchestré par des **manifestes YAML** :
- OlaresManifest.yaml (métadonnées app)
- Deployment.yaml (Kubernetes)
- Service.yaml (réseau)
- PVC.yaml (stockage)

### 2. **BFL Gateway comme Routeur**

Le BFL Gateway d'Olares route automatiquement le trafic vers les applications via des **annotations** :

```yaml
metadata:
  annotations:
    applications.app.bytetrade.io/entrances: |
      - name: api
        host: api.getloodbrain.{username}.{olares.local}
        port: 8080
        title: Getlood Brain API
```

### 3. **Réseau Olares Natif**

Chaque application obtient automatiquement :
- Un **namespace Kubernetes** : `getloodbrain-{username}`
- Un **service DNS** : `getloodbrain-svc.getloodbrain-{username}.svc.cluster.local`
- Une **URL publique** : `https://api.getloodbrain.{username}.olares.local`

### 4. **Authentification via Authelia**

Toutes les requêtes passent par **Authelia** (SSO d'Olares) :
- Pas besoin de gérer des tokens
- Authentification unique
- RBAC intégré

---

## 📦 Applications Olares Rebrandées

### 1. Getlood Brain (AIOS Rebrandé)

**Rôle** : Kernel AI pour orchestration intelligente.

**Composants** :
- AIOS Scheduler
- AIOS Context Manager
- AIOS Memory Manager
- AIOS LLM Core
- AIOS Storage Manager

**Architecture de Communication** :

```
┌─────────────────────────────────────────────────────┐
│  Getlood Brain Architecture                        │
│                                                     │
│  ┌──────────────────────────────────────────────┐  │
│  │  API Gateway (Port 8080)                     │  │
│  │  - Entrée principale                         │  │
│  │  - Routing des requêtes                      │  │
│  └──────────────┬───────────────────────────────┘  │
│                 │                                   │
│     ┌───────────┼───────────┬───────────┐          │
│     ↓           ↓           ↓           ↓          │
│  ┌──────┐  ┌────────┐  ┌────────┐  ┌─────────┐   │
│  │Sched │  │Context │  │Memory  │  │LLM Core │   │
│  │:8001 │  │:8002   │  │:8003   │  │:8004    │   │
│  └──┬───┘  └────────┘  └───┬────┘  └────┬────┘   │
│     │                       │            │         │
│     └───────────────────────┼────────────┘         │
│                             ↓                      │
│                    ┌─────────────────┐             │
│                    │ Storage Manager │             │
│                    │      :8005      │             │
│                    └─────────────────┘             │
└─────────────────────────────────────────────────────┘
```

---

### 2. Getlood Agents (SmythOS Rebrandé)

**Rôle** : Runtime pour agents AI autonomes.

**Composants** :
- SmythOS Agent Runtime
- SmythOS LLM Manager
- SmythOS Workflow Engine
- SmythOS Tool Manager

**Architecture de Communication** :

```
┌─────────────────────────────────────────────────────┐
│  Getlood Agents Architecture                       │
│                                                     │
│  ┌──────────────────────────────────────────────┐  │
│  │  Agent Runtime (Port 8080)                   │  │
│  │  - Exécution des agents                      │  │
│  │  - Communication inter-agents                │  │
│  └──────────────┬───────────────────────────────┘  │
│                 │                                   │
│     ┌───────────┼───────────┬───────────┐          │
│     ↓           ↓           ↓           ↓          │
│  ┌──────┐  ┌────────┐  ┌────────┐  ┌──────────┐  │
│  │LLM   │  │Workflow│  │Tool    │  │UI        │  │
│  │Mgr   │  │Engine  │  │Manager │  │:3000     │  │
│  │:8081 │  │:8082   │  │:8083   │  │          │  │
│  └──┬───┘  └───┬────┘  └────────┘  └──────────┘  │
│     │          │                                   │
│     │          │   Calls Brain API                │
│     │          └──────────────────────────────────►│
│     │                                               │
│     │   Calls Getlood LLM                          │
│     └──────────────────────────────────────────────►│
└─────────────────────────────────────────────────────┘
```

---

### 3. Getlood LLM (Ollama Rebrandé)

**Rôle** : Runtime LLM local avec support GPU.

**Composants** :
- Ollama (LLM inference)
- Model Manager (téléchargement/gestion modèles)

**Architecture** :

```
┌─────────────────────────────────────────────────────┐
│  Getlood LLM Architecture                          │
│                                                     │
│  ┌──────────────────────────────────────────────┐  │
│  │  Ollama Runtime (Port 11434)                 │  │
│  │  - Inférence LLM                             │  │
│  │  - Gestion des modèles                       │  │
│  │  - Support GPU (optionnel)                   │  │
│  └──────────────────────────────────────────────┘  │
│                                                     │
│  ┌──────────────────────────────────────────────┐  │
│  │  Model Manager (Port 8080)                   │  │
│  │  - Téléchargement modèles                    │  │
│  │  - Mise à jour modèles                       │  │
│  └──────────────────────────────────────────────┘  │
│                                                     │
│  ┌──────────────────────────────────────────────┐  │
│  │  UI (Port 3000)                              │  │
│  │  - Interface de gestion                      │  │
│  └──────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

---

### 4. Getlood VectorDB (Qdrant Rebrandé)

**Rôle** : Base de données vectorielle pour embeddings.

**Architecture** :

```
┌─────────────────────────────────────────────────────┐
│  Getlood VectorDB Architecture                     │
│                                                     │
│  ┌──────────────────────────────────────────────┐  │
│  │  Qdrant (Port 6333 HTTP, 6334 gRPC)         │  │
│  │  - Stockage vectoriel                        │  │
│  │  - Recherche sémantique                      │  │
│  │  - Snapshots                                 │  │
│  └──────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

---

## 🔄 Communication entre Applications

### Via Services Kubernetes

Les applications communiquent via des **services Kubernetes** :

```yaml
# Getlood Agents appelle Getlood Brain
GET http://getloodbrain-svc.getloodbrain-{username}.svc:8080/api/scheduler/submit

# Getlood Brain appelle Getlood LLM
POST http://getloodllm-svc.getloodllm-{username}.svc:11434/api/generate

# Getlood Brain appelle Getlood VectorDB
POST http://getloodvectordb-svc.getloodvectordb-{username}.svc:6333/collections/search
```

### Via BFL Gateway (Externe)

Les utilisateurs accèdent via le **BFL Gateway** :

```bash
# Depuis le navigateur ou CLI
curl https://api.getloodbrain.alice.olares.local/api/scheduler/status
curl https://api.getloodagents.alice.olares.local/api/agents/list
curl https://api.getloodllm.alice.olares.local/api/tags
```

---

## 🚀 Déploiement

### Étape 1 : Créer les Namespaces

Olares crée automatiquement les namespaces lors de l'installation :

```bash
# Exemple pour utilisateur "alice"
getloodbrain-alice
getloodagents-alice
getloodllm-alice
getloodvectordb-alice
```

### Étape 2 : Déployer les Services

Chaque application déploie ses services :

```yaml
# Exemple pour Getlood Brain
apiVersion: v1
kind: Service
metadata:
  name: getloodbrain-svc
  namespace: getloodbrain-alice
spec:
  selector:
    app: getloodbrain
  ports:
    - name: api
      port: 8080
      targetPort: 8080
```

### Étape 3 : Configurer le BFL Gateway

Le BFL Gateway route automatiquement :

```yaml
# Route créée automatiquement
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: getloodbrain-ingress
  namespace: getloodbrain-alice
spec:
  rules:
    - host: api.getloodbrain.alice.olares.local
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: getloodbrain-svc
                port:
                  number: 8080
```

---

## 🔐 Sécurité

### Authentification

Toutes les requêtes passent par **Authelia** :

```
User Request
    ↓
BFL Gateway
    ↓
Authelia (SSO)
    ↓ (if authenticated)
Application
```

### Isolation

Chaque utilisateur a ses propres namespaces :

```bash
# Utilisateur Alice
getloodbrain-alice
getloodagents-alice

# Utilisateur Bob
getloodbrain-bob
getloodagents-bob
```

### Permissions

Les applications définissent leurs permissions via `sysData` :

```yaml
permission:
  sysData:
    - dataType: legacy_api
      appName: getloodbrain
      port: 8080
      group: api.getloodbrain
      version: v1
      ops:
        - POST
        - GET
        - PUT
        - DELETE
```

---

## 📊 Monitoring

### Via Kubernetes

```bash
# Voir tous les pods
kubectl get pods -A | grep getlood

# Voir les logs
kubectl logs -f {pod-name} -n {namespace}

# Voir les ressources
kubectl top pods -n {namespace}
```

### Via Olares Dashboard

```
Control Hub > Applications > Getlood *
- Status
- CPU/Memory usage
- Logs
- Metrics
```

---

## ✅ Avantages de cette Architecture

| Aspect | Avantage |
|---|---|
| **Simplicité** | Zéro code custom, uniquement YAML |
| **Native** | Utilise 100% l'infrastructure Olares |
| **Isolation** | Chaque app dans son namespace |
| **Sécurité** | Authelia SSO intégré |
| **Scalabilité** | Kubernetes natif |
| **Maintenance** | Sync avec upstream via Git |
| **UX** | Interface Olares familière |

---

## 🎯 Conclusion

Cette architecture **native Olares** est la plus simple et la plus élégante :

1. **Pas de MCP Gateway** : Le BFL Gateway d'Olares suffit
2. **Pas de code custom** : Uniquement des manifestes YAML
3. **Pas de modification d'Olares** : Respect de la contrainte
4. **Intégration parfaite** : AIOS et SmythOS deviennent des citoyens de première classe d'Olares

**Getlood OS** devient un **ensemble d'applications Olares** qui transforment la plateforme en un système d'exploitation AI complet.
