# autokueng-data
The api to upload images and other data and serve them via http for the autokueng-api

- [frontend](https://github.com/janlauber/autokueng-frontend)
- [backend api](https://github.com/janlauber/autokueng-api)

## Installation

### Env Variables
```bash
# must be the same as in autokueng backend api
export JWT_SECRET_KEY=<secret>
export URL=http://localhost:9000
```

### Persistent storage
The data is stored in the following directory:
- `/opt/autokueng-data`

### Docker
```bash
docker run -d \
--name autokueng-data \
--mount source=autokueng-data,target=/opt/autokueng-data \
-p 9000:9000 \
janlauber/autokueng-data:latest
```