package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"sync"
	"time"
)

const (
	// Tempo de vida do cache em segundos
	CACHE_TTL = 30 * time.Second
)

// getDecision é o wrapper principal
func (a *App) getDecision(userID, flagName string) (bool, error) {
	// 1. Obter os dados da flag (do cache ou dos serviços)
	info, err := a.getCombinedFlagInfo(flagName)
	if err != nil {
		return false, err
	}

	// 2. Executar a lógica de avaliação
	return a.runEvaluationLogic(info, userID), nil
}

// getCombinedFlagInfo busca os dados no Redis, com fallback para os microsserviços
func (a *App) getCombinedFlagInfo(flagName string) (*CombinedFlagInfo, error) {
	cacheKey := fmt.Sprintf("flag_info:%s", flagName)

	// 1. Tentar buscar do Cache (Redis)
	val, err := a.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		// Cache HIT
		var info CombinedFlagInfo
		if err := json.Unmarshal([]byte(val), &info); err == nil {
			log.Printf("Cache HIT para flag '%s'", flagName)
			return &info, nil
		}
		// Se o unmarshal falhar, trata como cache miss
		log.Printf("Erro ao desserializar cache para flag '%s': %v", flagName, err)
	}

	log.Printf("Cache MISS para flag '%s'", flagName)
	// 2. Cache MISS - Buscar dos serviços
	info, err := a.fetchFromServices(flagName)
	if err != nil {
		return nil, err
	}

	// 3. Salvar no Cache
	jsonData, _ := json.Marshal(info)
	if err := a.RedisClient.Set(ctx, cacheKey, jsonData, CACHE_TTL).Err(); err != nil {
		log.Printf("Erro ao salvar no cache para flag '%s': %v", flagName, err)
	}

	return info, nil
}

// fetchFromServices busca dados do flag-service e targeting-service concorrentemente
func (a *App) fetchFromServices(flagName string) (*CombinedFlagInfo, error) {
	var wg sync.WaitGroup
	wg.Add(2)

	var flagInfo *Flag
	var ruleInfo *TargetingRule
	var flagErr, ruleErr error

	// Goroutine 1: Buscar do flag-service
	go func() {
		defer wg.Done()
		flagInfo, flagErr = a.fetchFlag(flagName)
	}()

	// Goroutine 2: Buscar do targeting-service
	go func() {
		defer wg.Done()
		ruleInfo, ruleErr = a.fetchRule(flagName)
	}()

	wg.Wait()

	if flagErr != nil {
		return nil, flagErr
	}
	if ruleErr != nil {
		log.Printf("Aviso: Nenhuma regra de segmentação encontrada para '%s'. Usando padrão.", flagName)
	}

	return &CombinedFlagInfo{
		Flag: flagInfo,
		Rule: ruleInfo,
	}, nil
}

// // fetchFlag (função helper)
// func (a *App) fetchFlag(flagName string) (*Flag, error) {
// 	url := fmt.Sprintf("%s/flags/%s", a.FlagServiceURL, flagName)

// 	apiKey := os.Getenv("SERVICE_API_KEY")
// 	req, _ := http.NewRequest("GET", url, nil)
// 	req.Header.Set("Authorization", "Bearer "+apiKey)

// 	resp, err := a.HttpClient.Do(req)
// 	if err != nil {
// 		return nil, fmt.Errorf("erro ao chamar flag-service: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode == http.StatusNotFound {
// 		return nil, &NotFoundError{flagName}
// 	}
// 	if resp.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("flag-service retornou status %d", resp.StatusCode)
// 	}

// 	body, _ := io.ReadAll(resp.Body)
// 	var flag Flag
// 	if err := json.Unmarshal(body, &flag); err != nil {
// 		return nil, fmt.Errorf("erro ao desserializar resposta do flag-service: %w", err)
// 	}
// 	return &flag, nil
// }

var validFlagName = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func (a *App) fetchFlag(flagName string) (*Flag, error) {
	// 1. Validação do flagName
	if !validFlagName.MatchString(flagName) {
		return nil, fmt.Errorf("flagName inválido")
	}

	// 2. Parse seguro da URL base
	baseURL, err := url.Parse(a.FlagServiceURL)
	if err != nil {
		return nil, fmt.Errorf("URL base inválida: %w", err)
	}

	// 3. Construção segura do path
	baseURL.Path = path.Join(baseURL.Path, "flags", flagName)
	urlStr := baseURL.String()

	// 4. Recuperação da API Key
	apiKey := os.Getenv("SERVICE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("SERVICE_API_KEY não definida")
	}

	// 5. Criação da requisição com tratamento de erro
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// 6. Execução da requisição
	resp, err := a.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao chamar flag-service: %w", err)
	}
	defer resp.Body.Close()

	// 7. Tratamento dos status HTTP
	if resp.StatusCode == http.StatusNotFound {
		return nil, &NotFoundError{flagName}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("flag-service retornou status %d", resp.StatusCode)
	}

	// 8. Leitura e parsing do corpo da resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %w", err)
	}
	var flag Flag
	if err := json.Unmarshal(body, &flag); err != nil {
		return nil, fmt.Errorf("erro ao desserializar resposta do flag-service: %w", err)
	}
	return &flag, nil
}

// type App struct {
// 	TargetingServiceURL string
// 	HttpClient          *http.Client
// }

func (a *App) fetchRule(flagName string) (*TargetingRule, error) {
	// 1. Validação do flagName
	if !validFlagName.MatchString(flagName) {
		return nil, fmt.Errorf("flagName inválido")
	}

	// 2. Parse seguro da URL base
	baseURL, err := url.Parse(a.TargetingServiceURL)
	if err != nil {
		return nil, fmt.Errorf("URL base inválida: %w", err)
	}

	// 3. Construção segura do path
	baseURL.Path = path.Join(baseURL.Path, "rules", flagName)
	urlStr := baseURL.String()

	// 4. Recuperação da API Key
	apiKey := os.Getenv("SERVICE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("SERVICE_API_KEY não definida")
	}

	// 5. Criação da requisição com tratamento de erro
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// 6. Execução da requisição
	resp, err := a.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao chamar targeting-service: %w", err)
	}
	defer resp.Body.Close()

	// 7. Tratamento dos status HTTP
	if resp.StatusCode == http.StatusNotFound {
		return nil, &NotFoundError{flagName}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("targeting-service retornou status %d", resp.StatusCode)
	}

	// 8. Leitura e parsing do corpo da resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %w", err)
	}
	var rule TargetingRule
	if err := json.Unmarshal(body, &rule); err != nil {
		return nil, fmt.Errorf("erro ao desserializar resposta do targeting-service: %w", err)
	}
	return &rule, nil
}

// func (a *App) fetchRule(flagName string) (*TargetingRule, error) {
// 	url := fmt.Sprintf("%s/rules/%s", a.TargetingServiceURL, flagName)
// 	apiKey := os.Getenv("SERVICE_API_KEY") // Usa a mesma chave
// 	req, _ := http.NewRequest("GET", url, nil)
// 	req.Header.Set("Authorization", "Bearer "+apiKey)

// 	resp, err := a.HttpClient.Do(req)
// 	if err != nil {
// 		return nil, fmt.Errorf("erro ao chamar targeting-service: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode == http.StatusNotFound {
// 		return nil, &NotFoundError{flagName} // Não é um erro fatal
// 	}
// 	if resp.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("targeting-service retornou status %d", resp.StatusCode)
// 	}

// 	body, _ := io.ReadAll(resp.Body)
// 	var rule TargetingRule
// 	if err := json.Unmarshal(body, &rule); err != nil {
// 		return nil, fmt.Errorf("erro ao desserializar resposta do targeting-service: %w", err)
// 	}
// 	return &rule, nil
// }

// runEvaluationLogic é onde a decisão é tomada
func (a *App) runEvaluationLogic(info *CombinedFlagInfo, userID string) bool {
	if info.Flag == nil || !info.Flag.IsEnabled {
		return false
	}

	if info.Rule == nil || !info.Rule.IsEnabled {
		return true
	}

	// 3. Processa a regra (só temos "PERCENTAGE" por enquanto)
	rule := info.Rule.Rules
	if rule.Type == "PERCENTAGE" {
		// Converte o 'value' (que é interface{}) para float64
		percentage, ok := rule.Value.(float64)
		if !ok {
			log.Printf("Erro: valor da regra de porcentagem não é um número para a flag '%s'", info.Flag.Name)
			return false
		}

		// Calcula o "bucket" do usuário (0-99)
		userBucket := getDeterministicBucket(userID + info.Flag.Name)

		if float64(userBucket) < percentage {
			return true
		}
	}

	return false
}

func getDeterministicBucket(input string) int {
	// Usamos SHA256 (mais seguro) e pegamos os primeiros 4 bytes
	hasher := sha256.New()
	hasher.Write([]byte(input))
	hash := hasher.Sum(nil)

	// Converte 4 bytes para um uint32
	val := binary.BigEndian.Uint32(hash[:4])

	// Retorna o módulo 100
	return int(val % 100)
}
