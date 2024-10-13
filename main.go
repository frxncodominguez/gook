package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Output struct {
	URL       string                 `json:"url"`
	Condition string                 `json:"condition"`
	Template  map[string]interface{} `json:"template"`
}

type Webhook struct {
	Name    string   `json:"name"`
	Path    string   `json:"path"`
	Outputs []Output `json:"outputs"`
}

type Config struct {
	Webhooks []Webhook `json:"webhooks"`
}

func loadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening config file: %w", err)
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("decoding config: %w", err)
	}

	return &config, nil
}

func checkDuplicatePaths(webhooks []Webhook) error {
	seen := make(map[string]struct{})
	for _, webhook := range webhooks {
		if _, exists := seen[webhook.Path]; exists {
			return fmt.Errorf("duplicate path found: %s", webhook.Path)
		}
		seen[webhook.Path] = struct{}{}
	}
	return nil
}

func executeTemplate(templateData map[string]interface{}, tmpl map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for key, value := range tmpl {
		t, err := template.New(key).Parse(value.(string))
		if err != nil {
			return nil, fmt.Errorf("parsing template %s: %w", key, err)
		}
		var buf bytes.Buffer
		if err := t.Execute(&buf, templateData); err != nil {
			return nil, fmt.Errorf("executing template %s: %w", key, err)
		}
		result[key] = buf.String()
	}
	return result, nil
}

func evaluateCondition(condition string, data map[string]interface{}) (bool, error) {
	t, err := template.New("condition").Parse(condition)
	if err != nil {
		return false, fmt.Errorf("parsing condition: %w", err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return false, fmt.Errorf("executing condition: %w", err)
	}
	return buf.String() == "true", nil
}

func handleWebhook(webhook Webhook, w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var wg sync.WaitGroup
	for _, output := range webhook.Outputs {
		wg.Add(1)
		go func(output Output) {
			defer wg.Done()
			processOutput(output, body)
		}(output)
	}
	wg.Wait()

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Webhook processed")
}

func processOutput(output Output, body map[string]interface{}) {
	conditionMet, err := evaluateCondition(output.Condition, body)
	if err != nil {
		log.Printf("Error evaluating condition for %s: %v", output.URL, err)
		return
	}

	if !conditionMet {
		return
	}

	alteredBody, err := executeTemplate(body, output.Template)
	if err != nil {
		log.Printf("Error processing template for %s: %v", output.URL, err)
		return
	}

	alteredBodyJSON, err := json.Marshal(alteredBody)
	if err != nil {
		log.Printf("Error processing JSON for %s: %v", output.URL, err)
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(output.URL, "application/json", bytes.NewBuffer(alteredBodyJSON))
	if err != nil {
		log.Printf("Error posting to %s: %v", output.URL, err)
		return
	}
	defer resp.Body.Close()

	log.Printf("Posted to %s with status: %s", output.URL, resp.Status)
}

func main() {
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	if err := checkDuplicatePaths(config.Webhooks); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	mux := http.NewServeMux()
	for _, webhook := range config.Webhooks {
		webhook := webhook
		mux.HandleFunc(webhook.Path, func(w http.ResponseWriter, r *http.Request) {
			handleWebhook(webhook, w, r)
		})
	}

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		log.Println("Server is listening on port 8080...")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	log.Println("Server stopped gracefully")
}
