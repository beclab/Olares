# Getlood Agents

**Autonomous AI agents runtime powered by SmythOS**

Getlood Agents est le runtime pour agents AI autonomes, basé sur SmythOS. Il permet de créer, déployer et orchestrer des agents AI.

## Composants

### SmythOS Agent Runtime
Environnement d'exécution pour agents AI avec support multi-agents et communication inter-agents.

### SmythOS LLM Manager
Gestionnaire de modèles LLM pour distribuer les charges et optimiser les ressources.

### SmythOS Workflow Engine
Moteur d'orchestration de workflows pour créer des pipelines AI complexes.

### SmythOS Tool Manager
Gestionnaire d'outils et APIs pour étendre les capacités des agents.

## Installation

```bash
# Via Olares Market
Market > Search "Getlood Agents" > Install

# Via CLI
kubectl apply -f OlaresManifest.yaml
```

## Configuration

### Ressources Minimales

- CPU: 1000m (1 core)
- Memory: 2Gi
- Disk: 20Gi

### Variables d'Environnement

```yaml
BRAIN_API_URL: "http://getloodbrain-api:8080"
LLM_API_URL: "http://getloodllm-svc:11434"
LLM_MANAGER_URL: "http://getloodagents-llm-manager:8081"
WORKFLOW_ENGINE_URL: "http://getloodagents-workflow:8082"
```

## Dépendances

- **Getlood Brain** : Pour l'orchestration intelligente
- **Getlood LLM** : Pour l'inférence LLM

## URLs

- API: `https://api.getloodagents.{username}.olares.local`
- UI: `https://ui.getloodagents.{username}.olares.local`

## API Endpoints

### List Agents

```bash
curl https://api.getloodagents.{username}.olares.local/agents
```

### Create Agent

```bash
curl -X POST https://api.getloodagents.{username}.olares.local/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Research Agent",
    "description": "Agent for research tasks",
    "model": "llama3.1:8b",
    "system_prompt": "You are a research assistant.",
    "tools": ["web_search", "summarize"]
  }'
```

### Execute Agent

```bash
curl -X POST https://api.getloodagents.{username}.olares.local/agents/{agent_id}/execute \
  -H "Content-Type: application/json" \
  -d '{
    "input": "Research the latest AI trends",
    "context": {}
  }'
```

### Create Workflow

```bash
curl -X POST https://api.getloodagents.{username}.olares.local/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Research & Summarize",
    "steps": [
      {"agent": "research_agent", "action": "search"},
      {"agent": "summary_agent", "action": "summarize"}
    ]
  }'
```

## Monitoring

```bash
# Vérifier les pods
kubectl get pods -n getloodagents-{username}

# Voir les logs
kubectl logs -f getloodagents-runtime-0 -n getloodagents-{username}

# Vérifier les ressources
kubectl top pod -n getloodagents-{username}
```

## Exemples d'Agents

### Agent de Recherche

```json
{
  "name": "Web Researcher",
  "model": "llama3.1:8b",
  "system_prompt": "You are a web researcher. Search and analyze information.",
  "tools": ["web_search", "extract_text", "summarize"]
}
```

### Agent de Code

```json
{
  "name": "Code Assistant",
  "model": "deepseek-coder:6.7b",
  "system_prompt": "You are a coding assistant. Write clean, efficient code.",
  "tools": ["code_analyzer", "linter", "test_generator"]
}
```

### Agent de Traduction

```json
{
  "name": "Translator",
  "model": "mistral:7b",
  "system_prompt": "You are a professional translator.",
  "tools": ["detect_language", "translate"]
}
```

## Dépannage

### Problème: Agent ne répond pas

```bash
# Vérifier LLM
curl https://api.getloodllm.{username}.olares.local/api/tags

# Vérifier les logs
kubectl logs getloodagents-runtime-0 -n getloodagents-{username}
```

### Problème: Workflow échoue

```bash
# Vérifier Workflow Engine
kubectl logs getloodagents-workflow-0 -n getloodagents-{username}

# Redémarrer le workflow
kubectl delete pod getloodagents-workflow-0 -n getloodagents-{username}
```

## Documentation

- [SmythOS Documentation](https://smythos.com/docs)
- [Architecture Globale](../README.md)
- [Olares Documentation](https://docs.olares.com)

## Licence

Apache 2.0
