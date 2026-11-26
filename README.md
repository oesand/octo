# 🐙 Octo

**Octo** is a lightweight, code-generation-driven **Dependency Injection (DI)** framework for Go.  
It brings a compile-time–safe approach to dependency wiring — no reflection, no complex configuration, and no runtime performance penalty.

---

## ✨ Features

- ✅ Type-safe dependency injection using generics.
- 🧩 Automatic wiring of constructors and named injections.
- ⚡ Compile-time generation — zero runtime reflection.
- 🕹 Mediatr-style request and notification handler scanning.
- 🧠 Simple API with explicit, readable wiring code.

---

## 📦 Installation

Add **Octo** to your project:

```bash
go get github.com/oesand/octo
````

Install the **octogen** code generation tool:

```bash
go install github.com/oesand/octo/cmd/octogen@latest
```

This will install the `octogen` binary into your `$GOPATH/bin` (or `$HOME/go/bin` by default).

---

## 🚀 Quick Start

### 1️⃣ Define your structs and constructors

```go
package example

type Config struct {
    DSN string
}

type Repository struct {
    cfg *Config
}

func NewRepository(cfg *Config) *Repository {
    return &Repository{cfg: cfg}
}

type Service struct {
    repo *Repository
}

func NewService(repo *Repository) *Service {
    return &Service{repo: repo}
}
```

---

### 2️⃣ Mark dependencies with `octogen.Inject`

```go
package include

import "github.com/oesand/octo/octogen"

func IncludeAll() {
    // Struct injection
    octogen.Inject[*Config]()
    
    // Constructor-based injection
    octogen.Inject(NewRepository)
    octogen.Inject(NewService)
}
```

---

### 3️⃣ Generate dependency wiring

Run the generator:

```bash
octogen -gen
```

It produces a generated file:

```go
func IncludeAll(container *octo.Container) {
    octo.Inject(container, func(c *octo.Container) *Config {
        return &Config{ DSN: "postgres://..." }
    })

    octo.Inject(container, func(c *octo.Container) *Repository {
        return NewRepository(octo.Resolve[*Config](c))
    })

    octo.Inject(container, func(c *octo.Container) *Service {
        return NewService(octo.Resolve[*Repository](c))
    })
}
```

---

### 4️⃣ Use the container

```go
package main

import (
    "fmt"
    "github.com/oesand/octo"
    "yourapp/include"
)

func main() {
    container := octo.New()
    include.IncludeAll(container)

    service := octo.Resolve[*Service](container)
    fmt.Println("Service ready:", service)
}
```

---

## ⚙️ Mediatr Scanning Example

`ScanForMediatr` automatically discovers and injects all request and notification handlers:

```go
func IncludeHandlers() {
    octogen.ScanForMediatr()
}
```

Generates:

```go
func IncludeHandlers(container *octo.Container) {
    octo.TryInject(container, func(c *octo.Container) *user.ReqHandler {
        return &user.ReqHandler{
            Repo: octo.Resolve[*Repository](c),
        }
    })
}
```

---

## Dependency cycles

Circular dependencies are not supported or detected. Keep provider dependencies acyclic:
both a direct cycle (`A → A`) and an indirect cycle (`A → B → A`) can deadlock during
resolution instead of reporting an error. This also applies when mutually dependent
providers are initialized concurrently.

Break the cycle by moving shared functionality into a separate service or passing the
required data to a method instead of resolving the other service in its constructor.

---

## 🧪 Testing

```bash
go test ./...
```

---

## 🧰 Example Commands

Generate DI wiring and run your app:

```bash
octogen
go run ./cmd/myapp
```
