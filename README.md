# Wing

[![golang](https://img.shields.io/badge/golang-%E2%89%A51.11-blue.svg?style=flat-square)](https://golang.org/)

Application orchestration backed by Kubernetes.



#### Build

```bash
make
```

#### Build Docker Image

```bash
docker build -t wing:latest .
```



## Configuration

#### Kubernetes

```yaml
kubernetes: 
  namespace: default    # Kubernetes namespace to deploy application
  kubeConfig: ./config  # kubeconfig to access to kube-apiserver. Let it empty when Wing runs within Kubernetes cluster (Wing will search for in-cluser kubeconfig instead). 
```

#### Database

Wing saves application configurations, users, and RBAC roles to database. 

```yaml
database:
  dsn: "root:123456@tcp(127.0.0.1:3306)/wing?charset=utf8&parseTime=true"
  engine: mysql
```



## Development

#### Run backend

```bash
make && bin/wing -config config.yml -debug
```

Run frontend

```bash
cd dashboard
npm run dev
```

