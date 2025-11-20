# GoPolyAI 🚀

![Version](https://img.shields.io/badge/version-v1.0.0-blue.svg)
![Go](https://img.shields.io/badge/go-1.21%2B-00ADD8.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

**The Ultimate Vendor-Agnostic AI Gateway for Go.**

**GoPolyAI** is a high-performance, interface-based library that unifies multiple AI providers (OpenAI, Google Gemini, Anthropic Claude, Ollama/Local) under a single, standardized API. It eliminates vendor lock-in, simplifies switching between models, and ensures high availability with built-in fallback mechanisms.

> **Write your code once. Run it with any AI.**

---

## 🧐 Why GoPolyAI? (The Problem vs. The Solution)

Before GoPolyAI, switching from OpenAI to Google Gemini required rewriting HTTP clients, changing JSON structures, and handling different authentication headers.

| Feature | ❌ Without GoPolyAI (The Old Way) | ✅ With GoPolyAI (The Elite Way) |
| :--- | :--- | :--- |
| **Provider Switching** | Requires rewriting logic & structs. | **Zero code changes.** Just change a config string. |
| **Codebase** | Cluttered with `if provider == "openai"` blocks. | **Clean & Polymorphic.** One interface (`AIProvider`). |
| **Reliability** | If OpenAI goes down, your app crashes. | **Resilient.** Automatic fallback to secondary providers. |
| **Development Cost** | You pay for every test API call. | **Free.** Use `mock` or `ollama` for local dev. |
| **Learning Curve** | Must learn every provider's specific API. | Learn **one method**: `Generate()`. |

---

## 🌟 Key Features

* **🧩 True Polymorphism:** A single `AIProvider` interface abstracts away all complexity.
* **🛡️ Smart Fallback System:** Automatically switches to a backup provider (e.g., Google) if the primary one (e.g., OpenAI) fails or times out. **Zero downtime.**
* **🏠 Local & Cloud Hybrid:** seamless support for local LLMs via **Ollama** alongside cloud giants.
* **⚡ Factory Pattern Support:** dynamic provider selection via CLI flags or Environment Variables.
* **🧪 Built-in Mocking:** Includes a zero-cost Mock client for unit testing and UI development.
* **🐳 Docker Ready:** Designed to work flawlessly within containerized environments.

---

## 📦 Installation

Install the library into your Go project:

```bash
go get [github.com/AhmetTasdemir/gopolyai@v1.0.0](https://github.com/AhmetTasdemir/gopolyai@v1.0.0)
````

*(Ensure your `go.mod` is initialized. If not, run `go mod init myproject` first).*

-----

## 🚀 Quick Start Guide

This example demonstrates how to build a CLI tool that can switch between OpenAI, Google, and Local AI without changing the core logic.

### 1\. Create `main.go`

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"[github.com/AhmetTasdemir/gopolyai/pkg/ai](https://github.com/AhmetTasdemir/gopolyai/pkg/ai)"
	"[github.com/AhmetTasdemir/gopolyai/pkg/ai/anthropic](https://github.com/AhmetTasdemir/gopolyai/pkg/ai/anthropic)"
	"[github.com/AhmetTasdemir/gopolyai/pkg/ai/google](https://github.com/AhmetTasdemir/gopolyai/pkg/ai/google)"
	"[github.com/AhmetTasdemir/gopolyai/pkg/ai/ollama](https://github.com/AhmetTasdemir/gopolyai/pkg/ai/ollama)"
	"[github.com/AhmetTasdemir/gopolyai/pkg/ai/openai](https://github.com/AhmetTasdemir/gopolyai/pkg/ai/openai)"
)

func main() {
	// 1. Dynamic Configuration via CLI Flags
	provider := flag.String("p", "ollama", "Provider: openai, google, anthropic, ollama")
	apiKey := flag.String("k", os.Getenv("AI_API_KEY"), "API Key")
	model := flag.String("m", "", "Model name (optional)")
	flag.Parse()

	prompt := "Explain Quantum Computing in one sentence."
	if len(flag.Args()) > 0 {
		prompt = flag.Args()[0]
	}

	// 2. Factory Pattern: Select the implementation
	var client ai.AIProvider

	switch *provider {
	case "openai":
		client = openai.NewClient(*apiKey)
	case "google":
		client = google.NewClient(*apiKey)
	case "anthropic":
		client = anthropic.NewClient(*apiKey)
	case "ollama":
		client = ollama.NewClient() // No API key needed for local
	default:
		log.Fatalf("Unknown provider: %s", *provider)
	}

	// 3. Configure (Optional overrides)
	cfg := ai.Config{Temperature: 0.7}
	if *model != "" {
		cfg.ModelName = *model
	}
	client.Configure(cfg)

	// 4. Execution (The Polymorphic Magic)
	fmt.Printf("--- Using Provider: %s ---\n", client.Name())
	
	resp, err := client.Generate(context.Background(), prompt)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println(">> Response:", resp)
}
```

### 2\. Run It\!

**Scenario A: Free Local Development (Ollama)**
*Prerequisite: Install Ollama from ollama.com*

```bash
go run main.go -p ollama "Who is Alan Turing?"
```

**Scenario B: Production with OpenAI**

```bash
export AI_API_KEY="sk-your-openai-key"
go run main.go -p openai -m "gpt-4o" "Write a poem about Go."
```

**Scenario C: Switching to Google Gemini**

```bash
go run main.go -p google -k "AIza-your-google-key" "What is the capital of Turkey?"
```

-----

## 🛡️ Advanced Usage: High Availability (Fallback)

GoPolyAI shines in production environments where reliability is non-negotiable. Use the **Composite Pattern** to chain providers.

```go
package main

import (
    "context"
    "fmt"
    "[github.com/AhmetTasdemir/gopolyai/pkg/ai](https://github.com/AhmetTasdemir/gopolyai/pkg/ai)"
    "[github.com/AhmetTasdemir/gopolyai/pkg/ai/google](https://github.com/AhmetTasdemir/gopolyai/pkg/ai/google)"
    "[github.com/AhmetTasdemir/gopolyai/pkg/ai/openai](https://github.com/AhmetTasdemir/gopolyai/pkg/ai/openai)"
)

func main() {
    // Define Primary and Secondary providers
    primary := openai.NewClient("sk-...")
    secondary := google.NewClient("AIza-...")

    // Create the Fallback Client
    // If OpenAI fails (401, 500, Timeout), Google will automatically take over.
    resilientClient := ai.NewFallbackClient(primary, secondary)

    // Use it just like any other provider!
    resp, err := resilientClient.Generate(context.Background(), "Critical mission query")
    
    if err != nil {
        fmt.Println("Both systems failed!", err)
    } else {
        fmt.Println("Success:", resp)
        fmt.Println("Served by:", resilientClient.Name()) 
        // Output might be: "SmartFallback (Pri: OpenAI -> Sec: Google)"
    }
}
```

-----

## 🌍 Real-World Scenarios

### 1\. The "Cost-Efficient" Startup Pipeline

**Problem:** Using GPT-4 for development and CI/CD tests is too expensive.
**Solution with GoPolyAI:**

  * **Local/Dev Environment:** Set `AI_PROVIDER=ollama`. Developers use Llama3 locally for free.
  * **Staging Environment:** Set `AI_PROVIDER=openai` with `gpt-3.5-turbo` for cheap cloud testing.
  * **Production Environment:** Set `AI_PROVIDER=anthropic` with `claude-3.5-sonnet` for high-quality user responses.
  * *All without changing a single line of Go code.*

### 2\. The "Never-Down" Enterprise Service

**Problem:** Your chatbot relies on OpenAI. When OpenAI has an outage, your customers leave.
**Solution with GoPolyAI:**
Implement the `FallbackClient`. Set OpenAI as primary and Azure OpenAI or Google Gemini as secondary. Your service achieves 99.99% availability by diversifying dependencies.

### 3\. A/B Testing Models

**Problem:** You don't know if Claude 3 is better than GPT-4 for your specific use case.
**Solution with GoPolyAI:**
Write a simple loop that initializes both clients and sends the same prompt to both. Log the results and compare them instantly.

-----

## 📂 Supported Providers & Configuration

| Provider | Keyword | Auth Requirement | Default Model |
| :--- | :--- | :--- | :--- |
| **OpenAI** | `openai` | API Key | `gpt-3.5-turbo` |
| **Google** | `google` | API Key | `gemini-1.5-flash` |
| **Anthropic**| `anthropic`| API Key | `claude-3-5-sonnet`|
| **Ollama** | `ollama` | None (Localhost) | `llama3` |
| **Mock** | `mock` | None | N/A |

-----

## 🤝 Contributing

Contributions are welcome\!

1.  Fork the project.
2.  Create your feature branch (`git checkout -b feature/AmazingFeature`).
3.  Commit your changes (`git commit -m 'Add some AmazingFeature'`).
4.  Push to the branch (`git push origin feature/AmazingFeature`).
5.  Open a Pull Request.

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.

-----

**Built with ❤️ by [Ahmet Taşdemir](https://github.com/AhmetTasdemir)**
