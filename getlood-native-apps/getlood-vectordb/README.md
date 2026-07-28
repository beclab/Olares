# Getlood VectorDB

**High-performance vector database for AI embeddings powered by Qdrant**

Getlood VectorDB est la base de données vectorielle pour Olares, basée sur Qdrant. Elle stocke et recherche des embeddings pour RAG et autres applications AI.

## Composants

### Qdrant
Base de données vectorielle haute performance avec support de recherche sémantique.

## Installation

```bash
# Via Olares Market
Market > Search "Getlood VectorDB" > Install

# Via CLI
kubectl apply -f OlaresManifest.yaml
```

## Configuration

### Ressources Minimales

- CPU: 1000m (1 core)
- Memory: 2Gi
- Disk: 50Gi

### Variables d'Environnement

```yaml
QDRANT__SERVICE__HTTP_PORT: "6333"
QDRANT__SERVICE__GRPC_PORT: "6334"
QDRANT__STORAGE__STORAGE_PATH: "/qdrant/storage"
QDRANT__STORAGE__SNAPSHOTS_PATH: "/qdrant/storage/snapshots"
QDRANT__STORAGE__OPTIMIZERS__DEFAULT_SEGMENT_NUMBER: "4"
QDRANT__STORAGE__PERFORMANCE__MAX_SEARCH_THREADS: "4"
```

## URLs

- API: `https://api.getloodvectordb.{username}.olares.local`
- Dashboard: `https://dashboard.getloodvectordb.{username}.olares.local`

## API Endpoints

### Health Check

```bash
curl https://api.getloodvectordb.{username}.olares.local/
```

### List Collections

```bash
curl https://api.getloodvectordb.{username}.olares.local/collections
```

### Create Collection

```bash
curl -X PUT https://api.getloodvectordb.{username}.olares.local/collections/my_collection \
  -H "Content-Type: application/json" \
  -d '{
    "vectors": {
      "size": 384,
      "distance": "Cosine"
    }
  }'
```

### Insert Vectors

```bash
curl -X PUT https://api.getloodvectordb.{username}.olares.local/collections/my_collection/points \
  -H "Content-Type: application/json" \
  -d '{
    "points": [
      {
        "id": 1,
        "vector": [0.1, 0.2, 0.3, ...],
        "payload": {"text": "Hello world"}
      }
    ]
  }'
```

### Search Vectors

```bash
curl -X POST https://api.getloodvectordb.{username}.olares.local/collections/my_collection/points/search \
  -H "Content-Type: application/json" \
  -d '{
    "vector": [0.1, 0.2, 0.3, ...],
    "limit": 10
  }'
```

### Delete Collection

```bash
curl -X DELETE https://api.getloodvectordb.{username}.olares.local/collections/my_collection
```

## Exemples d'Utilisation

### RAG (Retrieval-Augmented Generation)

```python
import requests
import numpy as np

# 1. Créer une collection
requests.put(
    "https://api.getloodvectordb.{username}.olares.local/collections/knowledge_base",
    json={
        "vectors": {
            "size": 384,  # Dimension des embeddings
            "distance": "Cosine"
        }
    }
)

# 2. Insérer des documents (avec embeddings)
requests.put(
    "https://api.getloodvectordb.{username}.olares.local/collections/knowledge_base/points",
    json={
        "points": [
            {
                "id": 1,
                "vector": embedding_model.encode("Document 1 text"),
                "payload": {"text": "Document 1 text", "source": "doc1.pdf"}
            },
            # ... plus de documents
        ]
    }
)

# 3. Rechercher des documents pertinents
query_embedding = embedding_model.encode("User question")
results = requests.post(
    "https://api.getloodvectordb.{username}.olares.local/collections/knowledge_base/points/search",
    json={
        "vector": query_embedding.tolist(),
        "limit": 5
    }
).json()
```

### Recherche Sémantique

```bash
# 1. Créer collection pour articles
curl -X PUT https://api.getloodvectordb.{username}.olares.local/collections/articles \
  -H "Content-Type: application/json" \
  -d '{
    "vectors": {"size": 768, "distance": "Cosine"}
  }'

# 2. Insérer articles
# (générer embeddings avec un modèle comme sentence-transformers)

# 3. Rechercher articles similaires
curl -X POST https://api.getloodvectordb.{username}.olares.local/collections/articles/points/search \
  -H "Content-Type: application/json" \
  -d '{
    "vector": [...],  # embedding de la requête
    "limit": 10,
    "with_payload": true
  }'
```

## Monitoring

```bash
# Vérifier les pods
kubectl get pods -n getloodvectordb-{username}

# Voir les logs
kubectl logs -f getloodvectordb-qdrant-0 -n getloodvectordb-{username}

# Vérifier les ressources
kubectl top pod -n getloodvectordb-{username}
```

## Gestion des Collections

### Lister toutes les collections

```bash
curl https://api.getloodvectordb.{username}.olares.local/collections
```

### Obtenir les infos d'une collection

```bash
curl https://api.getloodvectordb.{username}.olares.local/collections/my_collection
```

### Créer un snapshot

```bash
curl -X POST https://api.getloodvectordb.{username}.olares.local/collections/my_collection/snapshots
```

### Lister les snapshots

```bash
curl https://api.getloodvectordb.{username}.olares.local/collections/my_collection/snapshots
```

## Dépannage

### Problème: Collection ne crée pas

```bash
# Vérifier l'espace disque
kubectl exec -it getloodvectordb-qdrant-0 -n getloodvectordb-{username} -- df -h

# Vérifier les logs
kubectl logs getloodvectordb-qdrant-0 -n getloodvectordb-{username}
```

### Problème: Recherche lente

```bash
# Vérifier le nombre de points dans la collection
curl https://api.getloodvectordb.{username}.olares.local/collections/my_collection

# Augmenter les ressources CPU
# Éditer OlaresManifest.yaml et augmenter les limites CPU
```

### Problème: Out of Memory

```bash
# Augmenter la limite mémoire
resources:
  limits:
    memory: 8Gi

# Redéployer
kubectl apply -f OlaresManifest.yaml
```

## Optimisation des Performances

### Indexation

```bash
# Optimiser les segments pour de meilleures performances
curl -X POST https://api.getloodvectordb.{username}.olares.local/collections/my_collection/index
```

### Nombre de Threads

Ajuster `MAX_SEARCH_THREADS` selon le nombre de CPU :

```yaml
QDRANT__STORAGE__PERFORMANCE__MAX_SEARCH_THREADS: "8"
```

## Documentation

- [Qdrant Documentation](https://qdrant.tech/documentation/)
- [Qdrant GitHub](https://github.com/qdrant/qdrant)
- [Architecture Globale](../README.md)
- [Olares Documentation](https://docs.olares.com)

## Licence

Apache 2.0
