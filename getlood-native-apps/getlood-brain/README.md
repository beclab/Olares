# Getlood Brain

**AI Kernel for intelligent orchestration and resource management**

Getlood Brain est le kernel AI d'Olares, basé sur AIOS (AI Operating System). Il orchestre intelligemment les ressources et les tâches AI.

## Composants

### AIOS Scheduler
Ordonnancement intelligent des tâches AI avec gestion des priorités et des ressources.

### AIOS Context Manager
Gestion du contexte conversationnel pour maintenir la cohérence entre les interactions.

### AIOS Memory Manager
Système de mémoire à long terme avec RAG (Retrieval-Augmented Generation) pour stocker et récupérer les connaissances.

### AIOS LLM Core
Adaptateur multi-modèles LLM pour supporter différents modèles (Ollama, vLLM, etc.).

### AIOS Storage Manager
Système de fichiers sémantique pour organiser et rechercher les données par leur signification.

## Installation

```bash
# Via Olares Market
Market > Search "Getlood Brain" > Install

# Via CLI
kubectl apply -f OlaresManifest.yaml
```

## Configuration

### Ressources Minimales

- CPU: 1000m (1 core)
- Memory: 2Gi
- Disk: 10Gi

### Variables d'Environnement

```yaml
AIOS_MODE: "scheduler"
CONTEXT_MANAGER_URL: "http://getloodbrain-context:8002"
MEMORY_MANAGER_URL: "http://getloodbrain-memory:8003"
LLM_CORE_URL: "http://getloodbrain-llm:8004"
STORAGE_MANAGER_URL: "http://getloodbrain-storage:8005"
```

## Dépendances

- **Getlood LLM** : Pour l'inférence LLM
- **Getlood VectorDB** : Pour le stockage des embeddings

## URLs

- API: `https://api.getloodbrain.{username}.olares.local`
- UI: `https://ui.getloodbrain.{username}.olares.local`

## API Endpoints

### Health Check

```bash
curl https://api.getloodbrain.{username}.olares.local/health
```

### Submit Task

```bash
curl -X POST https://api.getloodbrain.{username}.olares.local/api/scheduler/submit \
  -H "Content-Type: application/json" \
  -d '{
    "task": "generate_summary",
    "input": "Long text to summarize...",
    "priority": "high"
  }'
```

### Get Task Status

```bash
curl https://api.getloodbrain.{username}.olares.local/api/scheduler/status/{task_id}
```

## Monitoring

```bash
# Vérifier les pods
kubectl get pods -n getloodbrain-{username}

# Voir les logs
kubectl logs -f getloodbrain-scheduler-0 -n getloodbrain-{username}

# Vérifier les ressources
kubectl top pod -n getloodbrain-{username}
```

## Dépannage

### Problème: Scheduler ne démarre pas

```bash
# Vérifier les dépendances
kubectl get svc -n getloodllm-{username}
kubectl get svc -n getloodvectordb-{username}

# Vérifier les logs
kubectl logs getloodbrain-scheduler-0 -n getloodbrain-{username}
```

### Problème: Memory Manager erreur de connexion

```bash
# Vérifier Qdrant
curl https://api.getloodvectordb.{username}.olares.local/collections

# Redémarrer Memory Manager
kubectl delete pod getloodbrain-memory-0 -n getloodbrain-{username}
```

## Documentation

- [AIOS GitHub](https://github.com/agiresearch/AIOS)
- [Architecture Globale](../README.md)
- [Olares Documentation](https://docs.olares.com)

## Licence

Apache 2.0
