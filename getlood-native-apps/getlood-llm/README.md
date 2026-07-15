# Getlood LLM

**Local LLM runtime powered by Ollama with GPU acceleration**

Getlood LLM est le runtime LLM local pour Olares, basé sur Ollama. Il permet d'exécuter des modèles de langage localement avec support GPU.

## Composants

### Ollama
Runtime LLM haute performance pour l'inférence locale de modèles.

### Model Manager
Gestionnaire de modèles pour télécharger, mettre à jour et gérer les modèles LLM.

## Installation

```bash
# Via Olares Market
Market > Search "Getlood LLM" > Install

# Via CLI
kubectl apply -f OlaresManifest.yaml
```

## Configuration

### Ressources Minimales

- CPU: 2000m (2 cores)
- Memory: 4Gi
- Disk: 100Gi (pour stocker les modèles)
- GPU: Optionnel mais recommandé

### Support GPU

Pour activer le GPU :

```yaml
# Éditer OlaresManifest.yaml
resources:
  limits:
    nvidia.com/gpu: 1
```

### Variables d'Environnement

```yaml
OLLAMA_HOST: "0.0.0.0:11434"
OLLAMA_MODELS: "/root/.ollama/models"
OLLAMA_KEEP_ALIVE: "24h"
OLLAMA_NUM_PARALLEL: "4"
OLLAMA_MAX_LOADED_MODELS: "3"
```

## URLs

- API: `https://api.getloodllm.{username}.olares.local`
- UI: `https://ui.getloodllm.{username}.olares.local`

## Modèles Recommandés

### Llama 3.1 8B (4.7GB)
Excellent modèle généraliste, bon équilibre performance/qualité.

```bash
kubectl exec -it getloodllm-ollama-0 -n getloodllm-{username} -- ollama pull llama3.1:8b
```

### Mistral 7B (4.1GB)
Rapide et efficace, excellent pour les tâches générales.

```bash
kubectl exec -it getloodllm-ollama-0 -n getloodllm-{username} -- ollama pull mistral:7b
```

### Phi-3 Mini (2.3GB)
Léger et rapide, parfait pour les tâches simples.

```bash
kubectl exec -it getloodllm-ollama-0 -n getloodllm-{username} -- ollama pull phi3:mini
```

### DeepSeek Coder 6.7B (3.8GB)
Spécialisé pour le code, excellent pour la programmation.

```bash
kubectl exec -it getloodllm-ollama-0 -n getloodllm-{username} -- ollama pull deepseek-coder:6.7b
```

## API Endpoints

### List Models

```bash
curl https://api.getloodllm.{username}.olares.local/api/tags
```

### Generate Text

```bash
curl -X POST https://api.getloodllm.{username}.olares.local/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama3.1:8b",
    "prompt": "Explain quantum computing in simple terms.",
    "stream": false
  }'
```

### Chat Completion

```bash
curl -X POST https://api.getloodllm.{username}.olares.local/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama3.1:8b",
    "messages": [
      {"role": "user", "content": "Hello, who are you?"}
    ]
  }'
```

### Pull Model

```bash
curl -X POST https://api.getloodllm.{username}.olares.local/api/pull \
  -H "Content-Type: application/json" \
  -d '{
    "name": "llama3.1:8b"
  }'
```

### Delete Model

```bash
curl -X DELETE https://api.getloodllm.{username}.olares.local/api/delete \
  -H "Content-Type: application/json" \
  -d '{
    "name": "llama3.1:8b"
  }'
```

## Monitoring

```bash
# Vérifier les pods
kubectl get pods -n getloodllm-{username}

# Voir les logs
kubectl logs -f getloodllm-ollama-0 -n getloodllm-{username}

# Vérifier les ressources (incluant GPU)
kubectl top pod -n getloodllm-{username}
```

## Gestion des Modèles

### Lister les modèles installés

```bash
kubectl exec -it getloodllm-ollama-0 -n getloodllm-{username} -- ollama list
```

### Télécharger un modèle

```bash
kubectl exec -it getloodllm-ollama-0 -n getloodllm-{username} -- ollama pull llama3.1:8b
```

### Supprimer un modèle

```bash
kubectl exec -it getloodllm-ollama-0 -n getloodllm-{username} -- ollama rm llama3.1:8b
```

### Vérifier l'espace disque

```bash
kubectl exec -it getloodllm-ollama-0 -n getloodllm-{username} -- df -h /root/.ollama
```

## Dépannage

### Problème: Modèle ne charge pas

```bash
# Vérifier l'espace disque
kubectl exec -it getloodllm-ollama-0 -n getloodllm-{username} -- df -h

# Vérifier les logs
kubectl logs getloodllm-ollama-0 -n getloodllm-{username}
```

### Problème: Inférence lente

```bash
# Vérifier si GPU est utilisé
kubectl describe pod getloodllm-ollama-0 -n getloodllm-{username} | grep -A 5 "Limits"

# Augmenter les ressources CPU/Memory si nécessaire
```

### Problème: Out of Memory

```bash
# Réduire le nombre de modèles chargés
OLLAMA_MAX_LOADED_MODELS=1

# Augmenter la limite mémoire dans OlaresManifest.yaml
resources:
  limits:
    memory: 16Gi
```

## Documentation

- [Ollama Documentation](https://ollama.com)
- [Ollama GitHub](https://github.com/ollama/ollama)
- [Architecture Globale](../README.md)
- [Olares Documentation](https://docs.olares.com)

## Licence

Apache 2.0
