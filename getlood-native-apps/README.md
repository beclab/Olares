# Getlood Native Apps : Applications Olares Rebrandées

Ce répertoire contient les **4 applications Olares** qui transforment Olares en **Getlood OS**, une plateforme AI agentique complète.

## 🎯 Vue d'Ensemble

Les applications sont des **versions rebrandées** d'AIOS et SmythOS, intégrées nativement dans Olares via des manifestes YAML. Aucun code custom n'est nécessaire : tout est orchestré par Kubernetes, Docker et le BFL Gateway d'Olares.

## 📦 Applications

### 1. Getlood Brain (AIOS Rebrandé)

**Rôle** : Kernel AI pour orchestration intelligente des ressources.

**Composants** :
- AIOS Scheduler : Ordonnancement des tâches AI
- AIOS Context Manager : Gestion du contexte conversationnel
- AIOS Memory Manager : Mémoire à long terme avec RAG
- AIOS LLM Core : Adaptateur multi-modèles LLM
- AIOS Storage Manager : Système de fichiers sémantique

**URLs** :
- API : `https://api.getloodbrain.{username}.olares.local`
- UI : `https://ui.getloodbrain.{username}.olares.local`

**Dépendances** :
- Getlood LLM (pour l'inférence LLM)
- Getlood VectorDB (pour les embeddings)

---

### 2. Getlood Agents (SmythOS Rebrandé)

**Rôle** : Runtime pour agents AI autonomes.

**Composants** :
- SmythOS Agent Runtime : Exécution des agents
- SmythOS LLM Manager : Gestion des modèles LLM
- SmythOS Workflow Engine : Orchestration de workflows
- SmythOS Tool Manager : Gestion des outils et APIs

**URLs** :
- API : `https://api.getloodagents.{username}.olares.local`
- UI : `https://ui.getloodagents.{username}.olares.local`

**Dépendances** :
- Getlood Brain (pour l'orchestration)
- Getlood LLM (pour l'inférence LLM)

---

### 3. Getlood LLM (Ollama Rebrandé)

**Rôle** : Runtime LLM local avec support GPU.

**Composants** :
- Ollama : Inférence LLM locale
- Model Manager : Téléchargement et gestion des modèles

**URLs** :
- API : `https://api.getloodllm.{username}.olares.local`
- UI : `https://ui.getloodllm.{username}.olares.local`

**Modèles Recommandés** :
- Llama 3.1 8B (4.7GB) : Généraliste excellent
- Mistral 7B (4.1GB) : Rapide et efficace
- Phi-3 Mini (2.3GB) : Léger pour tâches simples
- DeepSeek Coder 6.7B (3.8GB) : Spécialisé code

**Dépendances** : Aucune

---

### 4. Getlood VectorDB (Qdrant Rebrandé)

**Rôle** : Base de données vectorielle pour embeddings.

**Composants** :
- Qdrant : Stockage et recherche vectorielle haute performance

**URLs** :
- API : `https://api.getloodvectordb.{username}.olares.local`
- Dashboard : `https://dashboard.getloodvectordb.{username}.olares.local`

**Dépendances** : Aucune

---

## 🚀 Installation

### Prérequis

- Une instance Olares fonctionnelle (v1.0.0+)
- Accès au Olares Market
- Au minimum 8GB RAM, 200GB disque

### Ordre d'Installation

1. **Getlood VectorDB** (infrastructure)
2. **Getlood LLM** (infrastructure)
3. **Getlood Brain** (kernel)
4. **Getlood Agents** (applications)

### Méthode 1 : Via Olares Market (Recommandée)

```bash
# 1. Ouvrir Olares Market
Market > Search "Getlood"

# 2. Installer dans l'ordre
Install "Getlood VectorDB" → Wait for "Running"
Install "Getlood LLM" → Wait for "Running"
Install "Getlood Brain" → Wait for "Running"
Install "Getlood Agents" → Wait for "Running"

# 3. Télécharger un modèle LLM
Open "Getlood LLM UI" → Download Model → "Llama 3.1 8B"
```

### Méthode 2 : Via CLI (Avancée)

```bash
# 1. Cloner ce dépôt
git clone https://github.com/Getlood/getlood-native-apps.git
cd getlood-native-apps

# 2. Installer les applications
kubectl apply -f getlood-vectordb/OlaresManifest.yaml
kubectl apply -f getlood-llm/OlaresManifest.yaml
kubectl apply -f getlood-brain/OlaresManifest.yaml
kubectl apply -f getlood-agents/OlaresManifest.yaml

# 3. Vérifier le statut
kubectl get applications -n user-space-{username}
```

---

## 🔧 Configuration

### GPU (Optionnel mais Recommandé)

Pour activer le GPU sur Getlood LLM :

```yaml
# Éditer getlood-llm/OlaresManifest.yaml
resources:
  limits:
    nvidia.com/gpu: 1  # Ajouter cette ligne
```

### Modèles LLM

Télécharger des modèles via l'UI ou CLI :

```bash
# Via CLI
kubectl exec -it getloodllm-ollama-0 -n getloodllm-{username} -- ollama pull llama3.1:8b
kubectl exec -it getloodllm-ollama-0 -n getloodllm-{username} -- ollama pull mistral:7b
```

### Mémoire et CPU

Ajuster les ressources selon votre hardware :

```yaml
# Dans chaque OlaresManifest.yaml
resources:
  requests:
    cpu: 1000m      # Minimum
    memory: 2Gi     # Minimum
  limits:
    cpu: 4000m      # Maximum
    memory: 8Gi     # Maximum
```

---

## 🧪 Tests

### Test 1 : Getlood VectorDB

```bash
curl https://api.getloodvectordb.{username}.olares.local/collections
# Attendu : {"result": {"collections": []}, "status": "ok", "time": 0.001}
```

### Test 2 : Getlood LLM

```bash
curl https://api.getloodllm.{username}.olares.local/api/tags
# Attendu : Liste des modèles installés
```

### Test 3 : Getlood Brain

```bash
curl https://api.getloodbrain.{username}.olares.local/health
# Attendu : {"status": "healthy", "components": {...}}
```

### Test 4 : Getlood Agents

```bash
curl https://api.getloodagents.{username}.olares.local/agents
# Attendu : {"agents": [], "total": 0}
```

### Test End-to-End

```bash
# Créer un agent simple
curl -X POST https://api.getloodagents.{username}.olares.local/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Agent",
    "description": "A simple test agent",
    "model": "llama3.1:8b",
    "system_prompt": "You are a helpful assistant."
  }'

# Exécuter l'agent
curl -X POST https://api.getloodagents.{username}.olares.local/agents/{agent_id}/execute \
  -H "Content-Type: application/json" \
  -d '{
    "input": "Hello, who are you?"
  }'
```

---

## 📊 Monitoring

### Via Olares Control Hub

```
Control Hub > Applications > Getlood *
- Status : Running / Stopped / Error
- CPU : Usage en temps réel
- Memory : Usage en temps réel
- Logs : Derniers 1000 lignes
```

### Via Kubernetes

```bash
# Pods
kubectl get pods -n getloodbrain-{username}
kubectl get pods -n getloodagents-{username}
kubectl get pods -n getloodllm-{username}
kubectl get pods -n getloodvectordb-{username}

# Logs
kubectl logs -f getloodbrain-scheduler-0 -n getloodbrain-{username}
kubectl logs -f getloodagents-runtime-0 -n getloodagents-{username}

# Services
kubectl get svc -n getloodbrain-{username}
```

---

## 🐛 Dépannage

### Problème : Application ne démarre pas

```bash
# Vérifier les événements
kubectl describe pod {pod-name} -n {namespace}

# Vérifier les logs
kubectl logs {pod-name} -n {namespace}

# Vérifier les ressources
kubectl top pod {pod-name} -n {namespace}
```

### Problème : Erreur 502 Bad Gateway

```bash
# Vérifier le service
kubectl get svc -n {namespace}

# Vérifier les endpoints
kubectl get endpoints -n {namespace}

# Redémarrer le pod
kubectl delete pod {pod-name} -n {namespace}
```

### Problème : Modèle LLM non trouvé

```bash
# Lister les modèles
kubectl exec -it getloodllm-ollama-0 -n getloodllm-{username} -- ollama list

# Télécharger un modèle
kubectl exec -it getloodllm-ollama-0 -n getloodllm-{username} -- ollama pull llama3.1:8b
```

### Problème : Manque de mémoire

```bash
# Augmenter les limites dans OlaresManifest.yaml
resources:
  limits:
    memory: 16Gi  # Au lieu de 8Gi

# Redéployer
kubectl apply -f {app}/OlaresManifest.yaml
```

---

## 🔄 Mise à Jour

### Via Olares Market

```
Market > My Apps > Getlood * > Update
```

### Via CLI

```bash
# Pull latest
git pull origin main

# Redéployer
kubectl apply -f getlood-vectordb/OlaresManifest.yaml
kubectl apply -f getlood-llm/OlaresManifest.yaml
kubectl apply -f getlood-brain/OlaresManifest.yaml
kubectl apply -f getlood-agents/OlaresManifest.yaml
```

---

## 🗑️ Désinstallation

### Via Olares Market

```
Market > My Apps > Getlood * > Uninstall
```

### Via CLI

```bash
# Supprimer les applications (dans l'ordre inverse)
kubectl delete application getloodagents -n user-space-{username}
kubectl delete application getloodbrain -n user-space-{username}
kubectl delete application getloodllm -n user-space-{username}
kubectl delete application getloodvectordb -n user-space-{username}

# Supprimer les namespaces
kubectl delete namespace getloodagents-{username}
kubectl delete namespace getloodbrain-{username}
kubectl delete namespace getloodllm-{username}
kubectl delete namespace getloodvectordb-{username}
```

---

## 📖 Documentation

- [Architecture Native Olares](./ARCHITECTURE.md)
- [Documentation AIOS](https://github.com/Getlood/AIOS)
- [Documentation SmythOS](https://smythos.com/docs)
- [Documentation Olares](https://docs.olares.com)

---

## 🤝 Contribution

Les contributions sont les bienvenues ! Voir [CONTRIBUTING.md](../CONTRIBUTING.md).

---

## 📄 Licence

Apache 2.0 - Voir [LICENSE](../LICENSE).

---

## 🙏 Remerciements

- [Olares](https://olares.com) pour la plateforme
- [AIOS](https://github.com/agiresearch/AIOS) pour le kernel AI
- [SmythOS](https://smythos.com) pour le runtime agents
- [Ollama](https://ollama.com) pour le runtime LLM
- [Qdrant](https://qdrant.tech) pour la base vectorielle
