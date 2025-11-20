# GoPolyAI 🚀 EN/TR

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

# ---------------------------------TÜRKÇE---------------------------------

# GoPolyAI 🚀

![Version](https://img.shields.io/badge/version-v1.0.0-blue.svg)
![Go](https://img.shields.io/badge/go-1.21%2B-00ADD8.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

**Go İçin Nihai Tedarikçi-Bağımsız (Vendor-Agnostic) AI Geçidi.**

**GoPolyAI**, birden fazla yapay zeka sağlayıcısını (OpenAI, Google Gemini, Anthropic Claude, Ollama/Local) tek bir standart API altında birleştiren, yüksek performanslı ve arayüz tabanlı bir kütüphanedir. Tedarikçi kilidini (vendor lock-in) ortadan kaldırır, modeller arası geçişi basitleştirir ve yerleşik yedekleme (fallback) mekanizmalarıyla yüksek erişilebilirlik sağlar.

> **Kodunu bir kez yaz. Herhangi bir AI ile çalıştır.**

---

## 🧐 Neden GoPolyAI? (Sorun vs. Çözüm)

GoPolyAI'dan önce, OpenAI'dan Google Gemini'ye geçmek; HTTP istemcilerini yeniden yazmayı, JSON yapılarını değiştirmeyi ve farklı kimlik doğrulama başlıklarını yönetmeyi gerektiriyordu.

| Özellik | ❌ GoPolyAI Olmadan (Eski Yöntem) | ✅ GoPolyAI İle (Elit Yöntem) |
| :--- | :--- | :--- |
| **Sağlayıcı Değiştirme** | Mantık ve struct yapılarının yeniden yazılmasını gerektirir. | **Sıfır kod değişikliği.** Sadece bir konfigürasyon metnini değiştirin. |
| **Kod Tabanı** | `if provider == "openai"` bloklarıyla doludur. | **Temiz ve Polimorfik.** Tek bir arayüz (`AIProvider`). |
| **Güvenilirlik** | OpenAI çökerse uygulamanız da çöker. | **Dayanıklı.** İkincil sağlayıcılara otomatik geçiş (Fallback). |
| **Geliştirme Maliyeti** | Her test API çağrısı için para ödersiniz. | **Ücretsiz.** Yerel geliştirme için `mock` veya `ollama` kullanın. |
| **Öğrenme Eğrisi** | Her sağlayıcının özel API'sini öğrenmek zorundasınız. | **Tek bir metodu** öğrenin: `Generate()`. |

---

## 🌟 Temel Özellikler

* **🧩 Gerçek Polimorfizm:** Tek bir `AIProvider` arayüzü tüm karmaşıklığı soyutlar.
* **🛡️ Akıllı Fallback (Yedekleme) Sistemi:** Birincil sağlayıcı (örn. OpenAI) başarısız olursa veya zaman aşımına uğrarsa otomatik olarak yedek sağlayıcıya (örn. Google) geçer. **Sıfır kesinti.**
* **🏠 Yerel & Bulut Hibrit Yapı:** Bulut devlerinin yanı sıra **Ollama** aracılığıyla yerel LLM'ler için kusursuz destek.
* **⚡ Factory Deseni Desteği:** CLI bayrakları veya Ortam Değişkenleri aracılığıyla dinamik sağlayıcı seçimi.
* **🧪 Dahili Mocklama:** Birim testleri ve UI geliştirme için sıfır maliyetli Mock istemcisi içerir.
* **🐳 Docker Uyumlu:** Konteynerli ortamlarda kusursuz çalışacak şekilde tasarlanmıştır.

---

## 📦 Kurulum

Kütüphaneyi Go projenize ekleyin:

```bash
go get [github.com/AhmetTasdemir/gopolyai@v1.0.0](https://github.com/AhmetTasdemir/gopolyai@v1.0.0)
````

*(Not: `go.mod` dosyanızın başlatıldığından emin olun. Değilse önce `go mod init projedi` komutunu çalıştırın).*

-----

## 🚀 Hızlı Başlangıç Rehberi

Bu örnek, çekirdek mantığı değiştirmeden OpenAI, Google ve Yerel AI arasında geçiş yapabilen bir CLI aracının nasıl oluşturulacağını gösterir.

### 1\. `main.go` Dosyasını Oluşturun

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
	// 1. CLI Bayrakları ile Dinamik Konfigürasyon
	provider := flag.String("p", "ollama", "Sağlayıcı: openai, google, anthropic, ollama")
	apiKey := flag.String("k", os.Getenv("AI_API_KEY"), "API Anahtarı")
	model := flag.String("m", "", "Model adı (opsiyonel)")
	flag.Parse()

	prompt := "Kuantum Bilgisayarları tek bir cümleyle açıkla."
	if len(flag.Args()) > 0 {
		prompt = flag.Args()[0]
	}

	// 2. Factory Deseni: Uygulamayı Seç
	var client ai.AIProvider

	switch *provider {
	case "openai":
		client = openai.NewClient(*apiKey)
	case "google":
		client = google.NewClient(*apiKey)
	case "anthropic":
		client = anthropic.NewClient(*apiKey)
	case "ollama":
		client = ollama.NewClient() // Yerel için API anahtarı gerekmez
	default:
		log.Fatalf("Bilinmeyen sağlayıcı: %s", *provider)
	}

	// 3. Konfigürasyon (Opsiyonel geçersiz kılmalar)
	cfg := ai.Config{Temperature: 0.7}
	if *model != "" {
		cfg.ModelName = *model
	}
	client.Configure(cfg)

	// 4. Çalıştırma (Polimorfik Sihir)
	fmt.Printf("--- Kullanılan Sağlayıcı: %s ---\n", client.Name())
	
	resp, err := client.Generate(context.Background(), prompt)
	if err != nil {
		log.Fatalf("Hata: %v", err)
	}

	fmt.Println(">> Cevap:", resp)
}
```

### 2\. Çalıştırın\!

**Senaryo A: Ücretsiz Yerel Geliştirme (Ollama)**
*Ön koşul: ollama.com adresinden Ollama'yı yükleyin.*

```bash
go run main.go -p ollama "Alan Turing kimdir?"
```

**Senaryo B: OpenAI ile Canlı Ortam (Production)**

```bash
export AI_API_KEY="sk-sizin-openai-anahtarınız"
go run main.go -p openai -m "gpt-4o" "Go dili hakkında bir şiir yaz."
```

**Senaryo C: Google Gemini'ye Geçiş**

```bash
go run main.go -p google -k "AIza-sizin-google-anahtarınız" "Türkiye'nin başkenti neresidir?"
```

-----

## 🛡️ Gelişmiş Kullanım: Yüksek Erişilebilirlik (Fallback)

GoPolyAI, güvenilirliğin pazarlık konusu olamayacağı üretim ortamlarında parlar. Sağlayıcıları zincirlemek için **Composite Pattern** kullanın.

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
    // Birincil ve İkincil sağlayıcıları tanımla
    primary := openai.NewClient("sk-...")
    secondary := google.NewClient("AIza-...")

    // Fallback İstemcisini Oluştur
    // OpenAI başarısız olursa (401, 500, Zaman Aşımı), Google otomatik olarak devreye girer.
    resilientClient := ai.NewFallbackClient(primary, secondary)

    // Diğer sağlayıcılar gibi kullanın!
    resp, err := resilientClient.Generate(context.Background(), "Kritik görev sorgusu")
    
    if err != nil {
        fmt.Println("Her iki sistem de başarısız oldu!", err)
    } else {
        fmt.Println("Başarılı:", resp)
        fmt.Println("Hizmet Veren:", resilientClient.Name()) 
        // Çıktı şöyle olabilir: "SmartFallback (Pri: OpenAI -> Sec: Google)"
    }
}
```

-----

## 🌍 Gerçek Hayat Senaryoları

### 1\. "Maliyet Etkin" Startup Süreci

**Sorun:** Geliştirme ve CI/CD testleri için GPT-4 kullanmak çok pahalıdır.
**GoPolyAI ile Çözüm:**

  * **Yerel/Dev Ortamı:** `AI_PROVIDER=ollama` ayarlayın. Geliştiriciler Llama3'ü yerelde ücretsiz kullanır.
  * **Staging (Test) Ortamı:** Ucuz bulut testi için `AI_PROVIDER=openai` ve `gpt-3.5-turbo` kullanın.
  * **Canlı (Production) Ortam:** Yüksek kaliteli kullanıcı yanıtları için `AI_PROVIDER=anthropic` ve `claude-3.5-sonnet` kullanın.
  * *Tüm bunlar tek bir satır Go kodu değiştirilmeden yapılır.*

### 2\. "Asla Kesilmeyen" Kurumsal Servis

**Sorun:** Chatbotunuz OpenAI'a güveniyor. OpenAI kesinti yaşadığında müşterileriniz hizmet alamıyor.
**GoPolyAI ile Çözüm:**
`FallbackClient` uygulayın. OpenAI'ı birincil, Azure OpenAI veya Google Gemini'yi ikincil olarak ayarlayın. Servisiniz bağımlılıkları çeşitlendirerek %99.99 kullanılabilirlik elde eder.

### 3\. Modellerin A/B Testi

**Sorun:** Claude 3'ün sizin kullanım durumunuz için GPT-4'ten daha iyi olup olmadığını bilmiyorsunuz.
**GoPolyAI ile Çözüm:**
Her iki istemciyi de başlatan ve aynı istemi ikisine de gönderen basit bir döngü yazın. Sonuçları günlüğe kaydedin ve anında karşılaştırın.

-----

## 📂 Desteklenen Sağlayıcılar ve Konfigürasyon

| Sağlayıcı | Anahtar Kelime | Kimlik Doğrulama | Varsayılan Model |
| :--- | :--- | :--- | :--- |
| **OpenAI** | `openai` | API Key | `gpt-3.5-turbo` |
| **Google** | `google` | API Key | `gemini-1.5-flash` |
| **Anthropic**| `anthropic`| API Key | `claude-3-5-sonnet`|
| **Ollama** | `ollama` | Yok (Localhost) | `llama3` |
| **Mock** | `mock` | Yok | N/A |

-----

## 🤝 Katkıda Bulunma

Katkılarınızı bekliyoruz\!

1.  Projeyi fork'layın.
2.  Kendi özellik dalınızı oluşturun (`git checkout -b feature/HarikaOzellik`).
3.  Değişikliklerinizi commit'leyin (`git commit -m 'HarikaOzellik eklendi'`).
4.  Dala push'layın (`git push origin feature/HarikaOzellik`).
5.  Bir Pull Request açın.

## 📄 Lisans

MIT Lisansı altında dağıtılmaktadır. Daha fazla bilgi için `LICENSE` dosyasına bakın.

-----

**[Ahmet Taşdemir](https://github.com/AhmetTasdemir) tarafından ❤️ ile geliştirildi.**
