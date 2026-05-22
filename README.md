# Imagens Docker para produção

Projeto de estudo com uma API HTTP em Go e três estratégias de Dockerfile: uma imagem **monolítica** (desenvolvimento ou referência) e duas imagens **multi-stage** otimizadas para produção (Distroless e Wolfi/Chainguard).

A API escuta na porta **8080** (`src/main.go`).

## Estrutura

```
.
├── src/
│   ├── main.go
│   └── go.mod
└── dockerfiles/
    ├── Dockerfile.standard    # imagem grande, com toolchain Go
    ├── Dockerfile.distroless  # multi-stage + runtime mínimo (Google)
    └── Dockerfile.wolfi       # multi-stage + runtime mínimo (Chainguard)
```

## Por que imagens eficientes em produção?

| Objetivo | Como as imagens “lean” ajudam |
|----------|-------------------------------|
| **Menor superfície de ataque** | A imagem final não carrega compilador, shell nem pacotes de build |
| **Deploy mais rápido** | Menos camadas e menos MB para pull no registry e nos nodes |
| **Menos CVEs** | Menos pacotes instalados = menos vulnerabilidades reportadas pelo scanner |
| **Startup previsível** | Binário estático (`CGO_ENABLED=0`) roda em bases `static`/`scratch` sem libc extra |

### Padrões usados nas imagens de produção

1. **Multi-stage build** — estágio `builder` compila; estágio final só copia o binário.
2. **Binário estático** — `CGO_ENABLED=0 GOOS=linux` para não depender de glibc no runtime.
3. **Runtime mínimo** — `gcr.io/distroless/static:nonroot` ou `cgr.dev/chainguard/static`.
4. **Usuário não-root** — Distroless `nonroot`; Chainguard segue o mesmo princípio de least privilege.

A imagem **standard** mantém a imagem completa `golang` no container final — útil para comparar tamanho e risco, **não** recomendada para produção.

## Comparativo dos Dockerfiles

| Dockerfile | Estágios | Imagem final (aprox.) | Produção? |
|------------|----------|------------------------|-----------|
| `Dockerfile.standard` | 1 | ~800 MB+ (toolchain Go) | Não |
| `Dockerfile.distroless` | 2 | ~2–5 MB (binário + base static) | Sim |
| `Dockerfile.wolfi` | 2 | Similar ao Distroless | Sim |

## Pré-requisitos

- [Docker](https://docs.docker.com/get-docker/) instalado e em execução
- Comandos executados na **raiz do repositório** (onde estão `src/` e `dockerfiles/`)

## Build das imagens

```bash
# Referência (monolítica) — não use em produção
docker build -f dockerfiles/Dockerfile.standard -t api:standard .

# Produção — Distroless (recomendado para estudo Google)
docker build -f dockerfiles/Dockerfile.distroless -t api:distroless .

# Produção — Wolfi/Chainguard
docker build -f dockerfiles/Dockerfile.wolfi -t api:wolfi .
```

Build com cache explícito (opcional, acelera rebuilds):

```bash
docker build -f dockerfiles/Dockerfile.distroless -t api:distroless --progress=plain .
```

Comparar tamanho das imagens locais:

```bash
docker images api:standard api:distroless api:wolfi
```

## Subir os containers

Mapeie a porta do host para **8080** no container:

```bash
# Distroless (produção)
docker run --rm -p 8080:8080 --name api-distroless api:distroless

# Wolfi (produção)
docker run --rm -p 8080:8080 --name api-wolfi api:wolfi

# Standard (apenas teste/comparação)
docker run --rm -p 8080:8080 --name api-standard api:standard
```

Em segundo plano:

```bash
docker run -d -p 8080:8080 --name api-distroless api:distroless
docker logs -f api-distroless
docker stop api-distroless
```

Testar se a API responde (outro terminal):

```bash
curl -v http://localhost:8080/
```

> Se a porta 8080 já estiver em uso no host, troque o mapeamento, por exemplo `-p 8081:8080`.

## Boas práticas adicionais (fora deste repo)

- **`.dockerignore`** — exclua `.git`, `README.md`, artefatos locais; só copie o necessário para o build.
- **Pin de tags** — em produção, fixe versões (`golang:1.25.0`, digest da base Distroless/Wolfi) em vez de `latest`.
- **Scan de vulnerabilidades** — `docker scout cves api:distroless` ou integração no CI.
- **Não commitar segredos** — use secrets do orchestrator (Kubernetes, Swarm) ou variáveis injetadas no runtime.

## Qual Dockerfile escolher?

- **Produção genérica:** `Dockerfile.distroless` — base amplamente usada, sem shell, usuário `nonroot`.
- **Supply chain hardened:** `Dockerfile.wolfi` — imagens Chainguard/Wolfi com foco em segurança e SBOM.
- **Aprendizado / baseline:** `Dockerfile.standard` — mostra o custo de não separar build e runtime.

## Referências

- [LinuxTips (YouTube)](https://youtu.be/CHIQjLSfjoM)
- [Multi-stage builds (Docker Docs)](https://docs.docker.com/build/building/multi-stage/)
- [Distroless images](https://github.com/GoogleContainerTools/distroless)
- [Chainguard Images](https://www.chainguard.dev/unchained/chainguard-images)
