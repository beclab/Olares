---
search: false
---
## Manage the Olares container

### Stop the container
To stop the running container:
```bash
docker stop oic
```

### Restart the container
To restart the container after it has been stopped:
```bash
docker start oic
```

### Uninstall the container
To completely remove the container and its associated data:
```bash
docker stop oic
docker rm oic
docker volume rm oic-data
```