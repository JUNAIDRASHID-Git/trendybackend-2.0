package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type GeminiHandler struct{}

func NewGeminiHandler() *GeminiHandler {
	return &GeminiHandler{}
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiGenerationConfig struct {
	ResponseMimeType string `json:"responseMimeType"`
}

type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func (h *GeminiHandler) GenerateDescription(c *gin.Context) {
	name := c.Query("name")
	nameAr := c.Query("name_ar")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product name is required"})
		return
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GEMINI_API_KEY environment variable is not configured on the backend"})
		return
	}

	// 1. Construct prompt targeted at Saudi Arabian hospitality and international premium standards
	prompt := fmt.Sprintf(
		"Generate 3 distinct, premium, mouth-watering product description choices for a food/beverage item named '%s' (alternative Arabic name: '%s'). "+
		"The descriptions must be optimized to attract both local Saudi Arabian customers (who value quality, hospitality, and local tastes) "+
		"and foreign expatriates/visitors (who love international premium standards). "+
		"For each choice, provide two descriptions: one in English (around 2-3 sentences) and one in Arabic (around 2-3 sentences). "+
		"You must return your output strictly in JSON format matching this schema: "+
		"{\"choices\": ["+
		"  {\"description_en\": \"English description option 1\", \"description_ar\": \"Arabic description option 1\"},"+
		"  {\"description_en\": \"English description option 2\", \"description_ar\": \"Arabic description option 2\"},"+
		"  {\"description_en\": \"English description option 3\", \"description_ar\": \"Arabic description option 3\"}"+
		"]}",
		name, nameAr,
	)

	// 2. Prepare request payload
	reqPayload := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			ResponseMimeType: "application/json",
		},
	}

	reqBytes, err := json.Marshal(reqPayload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize request: " + err.Error()})
		return
	}

	// 3. Post to Gemini API
	apiURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=" + apiKey
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call Gemini API: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read Gemini API response: " + err.Error()})
		return
	}

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   fmt.Sprintf("Gemini API returned status %d", resp.StatusCode),
			"details": string(respBytes),
		})
		return
	}

	// 4. Parse response
	var geminiResp geminiResponse
	if err := json.Unmarshal(respBytes, &geminiResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Gemini API response: " + err.Error()})
		return
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Empty response candidates returned from Gemini"})
		return
	}

	rawText := geminiResp.Candidates[0].Content.Parts[0].Text

	// Parse rawText to verify it is valid JSON and clean it
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(rawText), &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to parse generated description as structured JSON: " + err.Error(),
			"rawText": rawText,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *GeminiHandler) GenerateSEO(c *gin.Context) {
	name := c.Query("name")
	description := c.Query("description")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product name is required"})
		return
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GEMINI_API_KEY environment variable is not configured on the backend"})
		return
	}

	prompt := fmt.Sprintf(
		"Generate optimized SEO metadata and professional SEO insights for a product named '%s' with description '%s'. "+
		"You must generate: "+
		"1. meta_title: A catchy, search-optimized title under 60 characters. "+
		"2. meta_description: An engaging description for search results under 160 characters. "+
		"3. keywords: 5 to 8 comma-separated relevant search keywords. "+
		"4. slug: A URL-friendly, lowercase, alphanumeric-and-hyphen-only string. "+
		"5. seo_insight: A professional-grade, 2-3 sentence strategic explanation analyzing the target search intent, why these keywords/meta-data are highly effective, and expert recommendations for catalog ranking. "+
		"You must return your output strictly in JSON format matching this schema: "+
		"{\"meta_title\": \"SEO Meta Title\", \"meta_description\": \"SEO Meta Description\", \"keywords\": \"keyword1, keyword2, keyword3\", \"slug\": \"product-slug\", \"seo_insight\": \"SEO Insight Analysis\"}",
		name, description,
	)

	reqPayload := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			ResponseMimeType: "application/json",
		},
	}

	reqBytes, err := json.Marshal(reqPayload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize request: " + err.Error()})
		return
	}

	apiURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=" + apiKey
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call Gemini API: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read Gemini API response: " + err.Error()})
		return
	}

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   fmt.Sprintf("Gemini API returned status %d", resp.StatusCode),
			"details": string(respBytes),
		})
		return
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(respBytes, &geminiResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Gemini API response: " + err.Error()})
		return
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Empty response candidates returned from Gemini"})
		return
	}

	rawText := geminiResp.Candidates[0].Content.Parts[0].Text

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(rawText), &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to parse generated SEO as structured JSON: " + err.Error(),
			"rawText": rawText,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *GeminiHandler) GeneratePricingInsight(c *gin.Context) {
	name := c.Query("name")
	costStr := c.Query("cost")
	currentSaleStr := c.Query("current_sale")
	category := c.Query("category")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product name is required"})
		return
	}

	if costStr == "" {
		costStr = "0.00"
	}
	if currentSaleStr == "" {
		currentSaleStr = "0.00"
	}
	if category == "" {
		category = "General"
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GEMINI_API_KEY environment variable is not configured on the backend"})
		return
	}

	prompt := fmt.Sprintf(
		"Analyze the pricing strategy for a food/beverage/restaurant item named '%s' in the category '%s'. "+
		"The product's cost price is %s and its current sale price is %s. "+
		"Generate a dynamic, premium pricing recommendation, including: "+
		"1. competitor_price: A realistic average competitor price for this type of product. "+
		"2. suggested_price: An optimized suggested sale price that maintains a healthy profit margin while being highly competitive. "+
		"3. suggested_regular_price: A suggested regular/original price (non-sale price) that aligns with the premium positioning. "+
		"4. volume_impact: An estimated percentage increase (e.g. 15 for 15%%) in sales volume if this suggested price is applied. "+
		"5. explanation: A concise, professional, engaging explanation of this insight (e.g., 'Competitors are listing this for AED 13.50. Adjusting your price to AED 13.25 could increase volume by 12%%.'). "+
		"You must return your output strictly in JSON format matching this schema: "+
		"{\"competitor_price\": 13.50, \"suggested_price\": 13.25, \"suggested_regular_price\": 15.00, \"volume_impact\": 12, \"explanation\": \"Competitors are listing this for...\"}",
		name, category, costStr, currentSaleStr,
	)

	reqPayload := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			ResponseMimeType: "application/json",
		},
	}

	reqBytes, err := json.Marshal(reqPayload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize request: " + err.Error()})
		return
	}

	apiURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=" + apiKey
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call Gemini API: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read Gemini API response: " + err.Error()})
		return
	}

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   fmt.Sprintf("Gemini API returned status %d", resp.StatusCode),
			"details": string(respBytes),
		})
		return
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(respBytes, &geminiResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Gemini API response: " + err.Error()})
		return
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Empty response candidates returned from Gemini"})
		return
	}

	rawText := geminiResp.Candidates[0].Content.Parts[0].Text

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(rawText), &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to parse generated pricing insight as structured JSON: " + err.Error(),
			"rawText": rawText,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *GeminiHandler) ScanProduct(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "GEMINI_API_KEY environment variable is not configured on the backend"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
		return
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read uploaded file"})
		return
	}

	base64Data := base64.StdEncoding.EncodeToString(fileBytes)

	mimeType := file.Header.Get("Content-Type")
	if mimeType == "" || mimeType == "application/octet-stream" {
		detected := http.DetectContentType(fileBytes)
		if strings.HasPrefix(detected, "image/") {
			mimeType = detected
		} else {
			ext := strings.ToLower(filepath.Ext(file.Filename))
			switch ext {
			case ".jpg", ".jpeg":
				mimeType = "image/jpeg"
			case ".png":
				mimeType = "image/png"
			case ".webp":
				mimeType = "image/webp"
			default:
				mimeType = "image/jpeg"
			}
		}
	}

	promptText := "Analyze this product image. Extract and generate appropriate details to auto-fill a new product form. " +
		"You must generate:\n" +
		"1. name: A premium marketing name in English.\n" +
		"2. name_ar: A premium marketing name in Arabic.\n" +
		"3. description: A clear, mouth-watering, search-optimized description in English (2-3 sentences).\n" +
		"4. description_ar: A clear, mouth-watering, search-optimized description in Arabic (2-3 sentences).\n" +
		"5. weight_volume: The estimated weight or volume of this product based on its packaging or typical size (e.g. '1 KG', '500 ml', '250 Gram', 'Piece').\n" +
		"6. meta_title: A search-engine optimized title for the product (under 60 characters).\n" +
		"7. meta_description: A search-engine optimized description for the product (under 160 characters).\n" +
		"8. keywords: A comma-separated list of 5-10 search keywords for this product.\n\n" +
		"Return your response strictly in JSON format matching this schema:\n" +
		"{\n" +
		"  \"name\": \"English Product Name\",\n" +
		"  \"name_ar\": \"Arabic Product Name\",\n" +
		"  \"description\": \"English description here...\",\n" +
		"  \"description_ar\": \"Arabic description here...\",\n" +
		"  \"weight_volume\": \"1 KG\",\n" +
		"  \"meta_title\": \"SEO Title\",\n" +
		"  \"meta_description\": \"SEO Description...\",\n" +
		"  \"keywords\": \"keyword1, keyword2\"\n" +
		"}"

	reqPayload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"inlineData": map[string]string{
							"mimeType": mimeType,
							"data":     base64Data,
						},
					},
					{
						"text": promptText,
					},
				},
			},
		},
		"generationConfig": map[string]string{
			"responseMimeType": "application/json",
		},
	}

	reqBytes, err := json.Marshal(reqPayload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize request"})
		return
	}

	var geminiFailed bool
	var result map[string]interface{}

	apiURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=" + apiKey
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		log.Printf("ScanProduct: Failed to contact Gemini API: %v. Attempting OpenAI fallback...", err)
		geminiFailed = true
	} else {
		defer resp.Body.Close()
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("ScanProduct: Failed to read Gemini response: %v. Attempting OpenAI fallback...", err)
			geminiFailed = true
		} else if resp.StatusCode != http.StatusOK {
			log.Printf("ScanProduct: Gemini API returned status %d. Response: %s. Attempting OpenAI fallback...", resp.StatusCode, string(respBytes))
			geminiFailed = true
		} else {
			var geminiResp struct {
				Candidates []struct {
					Content struct {
						Parts []struct {
							Text string `json:"text"`
						} `json:"parts"`
					} `json:"content"`
				} `json:"candidates"`
			}

			if err := json.Unmarshal(respBytes, &geminiResp); err != nil || len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
				log.Printf("ScanProduct: Failed to parse Gemini response: %v. Attempting OpenAI fallback...", err)
				geminiFailed = true
			} else {
				rawText := geminiResp.Candidates[0].Content.Parts[0].Text
				if err := json.Unmarshal([]byte(rawText), &result); err != nil {
					log.Printf("ScanProduct: Failed to parse structured response: %v. Attempting OpenAI fallback...", err)
					geminiFailed = true
				}
			}
		}
	}

	if geminiFailed {
		openaiKey := os.Getenv("OPENAI_API_KEY")
		if openaiKey == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gemini API failed and no OPENAI_API_KEY environment variable configured on backend"})
			return
		}
		log.Println("ScanProduct: Triggering OpenAI GPT-4o-mini fallback scan...")
		result, err = scanProductWithOpenAI(base64Data, mimeType, openaiKey)
		if err != nil {
			log.Printf("ScanProduct: OpenAI fallback scan also failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "AI Scan failed on both Gemini and OpenAI", "details": err.Error()})
			return
		}
		log.Println("ScanProduct: OpenAI fallback scan succeeded!")
	}

	c.JSON(http.StatusOK, result)
}

func scanProductWithOpenAI(base64Data string, mimeType string, apiKey string) (map[string]interface{}, error) {
	promptText := "Analyze this product image. Extract and generate appropriate details to auto-fill a new product form. " +
		"You must generate:\n" +
		"1. name: A premium marketing name in English.\n" +
		"2. name_ar: A premium marketing name in Arabic.\n" +
		"3. description: A clear, mouth-watering, search-optimized description in English (2-3 sentences).\n" +
		"4. description_ar: A clear, mouth-watering, search-optimized description in Arabic (2-3 sentences).\n" +
		"5. weight_volume: The estimated weight or volume of this product based on its packaging or typical size (e.g. '1 KG', '500 ml', '250 Gram', 'Piece').\n" +
		"6. meta_title: A search-engine optimized title for the product (under 60 characters).\n" +
		"7. meta_description: A search-engine optimized description for the product (under 160 characters).\n" +
		"8. keywords: A comma-separated list of 5-10 search keywords for this product.\n\n" +
		"Return your response strictly in JSON format matching this schema:\n" +
		"{\n" +
		"  \"name\": \"English Product Name\",\n" +
		"  \"name_ar\": \"Arabic Product Name\",\n" +
		"  \"description\": \"English description here...\",\n" +
		"  \"description_ar\": \"Arabic description here...\",\n" +
		"  \"weight_volume\": \"1 KG\",\n" +
		"  \"meta_title\": \"SEO Title\",\n" +
		"  \"meta_description\": \"SEO Description...\",\n" +
		"  \"keywords\": \"keyword1, keyword2\"\n" +
		"}"

	imageURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)

	reqPayload := map[string]interface{}{
		"model": "gpt-4o-mini",
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": promptText,
					},
					{
						"type": "image_url",
						"image_url": map[string]string{
							"url": imageURL,
						},
					},
				},
			},
		},
		"response_format": map[string]string{
			"type": "json_object",
		},
	}

	reqBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai chat api returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBytes, &openAIResp); err != nil || len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("failed to parse openai response: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(openAIResp.Choices[0].Message.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse assistant content: %v", err)
	}

	return result, nil
}

func (h *GeminiHandler) GeneratePhotoshoot(c *gin.Context) {
	// 1. Parse image from request
	file, err := c.FormFile("image")
	if err != nil {
		log.Printf("GeneratePhotoshoot: FormFile error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}

	// Also get name and provider if provided
	name := c.PostForm("name")
	provider := c.PostForm("provider")
	if provider == "" {
		provider = "gemini"
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	openaiKey := os.Getenv("OPENAI_API_KEY")

	// Ensure key is configured for the selected provider
	if provider == "chatgpt" {
		if openaiKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "OPENAI_API_KEY environment variable is not configured on the backend"})
			return
		}
	} else {
		if apiKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "GEMINI_API_KEY environment variable is not configured on the backend"})
			return
		}
	}

	// 2. Open and read file bytes
	src, err := file.Open()
	if err != nil {
		log.Printf("GeneratePhotoshoot: file.Open error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
		return
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		log.Printf("GeneratePhotoshoot: io.ReadAll error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read uploaded file"})
		return
	}

	// 3. Encode to base64
	base64Data := base64.StdEncoding.EncodeToString(fileBytes)

	// Determine MIME type robustly
	mimeType := file.Header.Get("Content-Type")
	log.Printf("GeneratePhotoshoot: uploaded file filename: %s, initial Content-Type: %s, provider: %s", file.Filename, mimeType, provider)
	if mimeType == "" || mimeType == "application/octet-stream" {
		detected := http.DetectContentType(fileBytes)
		log.Printf("GeneratePhotoshoot: Sniffed MIME type from bytes: %s", detected)
		if strings.HasPrefix(detected, "image/") {
			mimeType = detected
		} else {
			ext := strings.ToLower(filepath.Ext(file.Filename))
			switch ext {
			case ".jpg", ".jpeg":
				mimeType = "image/jpeg"
			case ".png":
				mimeType = "image/png"
			case ".webp":
				mimeType = "image/webp"
			case ".gif":
				mimeType = "image/gif"
			case ".heic":
				mimeType = "image/heic"
			case ".heif":
				mimeType = "image/heif"
			default:
				mimeType = "image/jpeg" // Fallback to jpeg
			}
			log.Printf("GeneratePhotoshoot: inferred Content-Type from extension: %s", mimeType)
		}
	}

	uploadsDir := os.Getenv("UPLOADS_DIR")
	if uploadsDir == "" {
		uploadsDir = "./uploads"
	}
	_ = os.MkdirAll(uploadsDir, os.ModePerm)

	var finalPrompt string
	var finalImageBytes []byte
	var finalImageMime string

	if provider == "chatgpt" {
		// --- ChatGPT (OpenAI) Pipeline ---
		// 1. Generate photoshoot prompt
		prompt, err := generatePromptWithOpenAI(base64Data, mimeType, name, openaiKey)
		if err != nil {
			log.Printf("GeneratePhotoshoot (ChatGPT): Prompt generation failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ChatGPT prompt generation failed: " + err.Error()})
			return
		}
		finalPrompt = prompt

		// 2. Generate photoshoot image using DALL-E 3
		log.Printf("GeneratePhotoshoot (ChatGPT): Calling DALL-E 3 with prompt: %s", finalPrompt)
		detectedMime, decBytes, err := generateImageWithOpenAI(finalPrompt, openaiKey)
		if err != nil {
			log.Printf("GeneratePhotoshoot (ChatGPT): Image generation failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ChatGPT image generation failed: " + err.Error()})
			return
		}
		finalImageBytes = decBytes
		finalImageMime = detectedMime
	} else {
		// --- Gemini (Google) Pipeline ---
		// 1. Construct prompt for gemini-2.5-flash
		promptText := "You are a professional photographer. Look at this product image. Write a highly detailed prompt (in English) to generate a commercial-grade 8k food photoshoot of this product in a clean studio setting on a solid white background, aspect ratio 1:1, beautiful ambient studio lighting. Focus on product branding, style, and texture. Return ONLY a JSON object matching this schema: {\"prompt\": \"your detailed photoshoot prompt\"}"
		if name != "" {
			promptText = fmt.Sprintf("You are a professional photographer. Look at this product image of '%s'. Write a highly detailed prompt (in English) to generate a commercial-grade 8k food photoshoot of this product in a clean studio setting on a solid white background, aspect ratio 1:1, beautiful ambient studio lighting. Focus on product branding, style, and texture. Return ONLY a JSON object matching this schema: {\"prompt\": \"your detailed photoshoot prompt\"}", name)
		}

		reqPayload := map[string]interface{}{
			"contents": []map[string]interface{}{
				{
					"parts": []map[string]interface{}{
						{
							"inlineData": map[string]string{
								"mimeType": mimeType,
								"data":     base64Data,
							},
						},
						{
							"text": promptText,
						},
					},
				},
			},
			"generationConfig": map[string]string{
				"responseMimeType": "application/json",
			},
		}

		reqBytes, err := json.Marshal(reqPayload)
		if err != nil {
			log.Printf("GeneratePhotoshoot (Gemini): json.Marshal error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize prompt request"})
			return
		}

		// 2. Call Gemini 2.5 Flash
		apiURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=" + apiKey
		resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(reqBytes))
		if err != nil {
			log.Printf("GeneratePhotoshoot (Gemini): HTTP Post to Gemini error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to contact Gemini API: " + err.Error()})
			return
		}
		defer resp.Body.Close()

		respBytes, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			log.Printf("GeneratePhotoshoot (Gemini): Gemini API returned status %d. Body: %s", resp.StatusCode, string(respBytes))
			var googleErr struct {
				Error struct {
					Message string `json:"message"`
				} `json:"error"`
			}
			errMsg := "Gemini API returned error"
			if err := json.Unmarshal(respBytes, &googleErr); err == nil && googleErr.Error.Message != "" {
				errMsg = fmt.Sprintf("Gemini: %s", googleErr.Error.Message)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg, "details": string(respBytes)})
			return
		}

		var geminiResp struct {
			Candidates []struct {
				Content struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
			} `json:"candidates"`
		}

		if err := json.Unmarshal(respBytes, &geminiResp); err != nil || len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
			log.Printf("GeneratePhotoshoot (Gemini): Failed to parse Gemini response: %v, body: %s", err, string(respBytes))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Gemini response"})
			return
		}

		rawText := geminiResp.Candidates[0].Content.Parts[0].Text
		var parsedPrompt struct {
			Prompt string `json:"prompt"`
		}
		_ = json.Unmarshal([]byte(rawText), &parsedPrompt)

		if parsedPrompt.Prompt == "" {
			parsedPrompt.Prompt = "A professional 8k food photoshoot of the product, clean studio setting on a solid white background, 1:1 aspect ratio, beautiful ambient studio lighting"
		}
		finalPrompt = parsedPrompt.Prompt

		// 3. Call Imagen 4
		imagenModel := "imagen-4.0-generate-001"
		imagenURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:predict?key=%s", imagenModel, apiKey)
		
		imagenPayload := map[string]interface{}{
			"instances": []map[string]interface{}{
				{"prompt": finalPrompt},
			},
			"parameters": map[string]interface{}{
				"sampleCount": 1,
				"aspectRatio": "1:1",
				"outputMimeType": "image/jpeg",
			},
		}

		imgReqBytes, _ := json.Marshal(imagenPayload)
		
		client := &http.Client{Timeout: 30 * time.Second}
		imgResp, imgErr := client.Post(imagenURL, "application/json", bytes.NewBuffer(imgReqBytes))
		if imgErr != nil {
			log.Printf("GeneratePhotoshoot (Gemini): Imagen API request failed: %v", imgErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Imagen API request failed: " + imgErr.Error()})
			return
		}
		defer imgResp.Body.Close()
		imgRespBytes, _ := io.ReadAll(imgResp.Body)
		
		if imgResp.StatusCode != http.StatusOK {
			log.Printf("GeneratePhotoshoot (Gemini): Imagen API failed with status %d: %s", imgResp.StatusCode, string(imgRespBytes))
			var googleErr struct {
				Error struct {
					Message string `json:"message"`
				} `json:"error"`
			}
			errMsg := "Imagen API returned error"
			if err := json.Unmarshal(imgRespBytes, &googleErr); err == nil && googleErr.Error.Message != "" {
				errMsg = fmt.Sprintf("Imagen: %s", googleErr.Error.Message)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg, "details": string(imgRespBytes)})
			return
		}

		var imagenResult struct {
			Predictions []struct {
				BytesBase64Encoded string `json:"bytesBase64Encoded"`
			} `json:"predictions"`
		}
		if err := json.Unmarshal(imgRespBytes, &imagenResult); err != nil || len(imagenResult.Predictions) == 0 {
			log.Printf("GeneratePhotoshoot (Gemini): Failed to parse Imagen API predictions: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Imagen response", "details": string(imgRespBytes)})
			return
		}

		decBytes, err := base64.StdEncoding.DecodeString(imagenResult.Predictions[0].BytesBase64Encoded)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode generated image bytes"})
			return
		}
		finalImageBytes = decBytes
		finalImageMime = "image/jpeg"
	}

	// 4. Save and return generated image (no template fallback)
	newFileName := fmt.Sprintf("ai_photoshoot_%d.jpg", time.Now().UnixNano())
	if finalImageMime != "" {
		if strings.Contains(strings.ToLower(finalImageMime), "png") {
			newFileName = fmt.Sprintf("ai_photoshoot_%d.png", time.Now().UnixNano())
		} else if strings.Contains(strings.ToLower(finalImageMime), "webp") {
			newFileName = fmt.Sprintf("ai_photoshoot_%d.webp", time.Now().UnixNano())
		}
	}
	targetPath := filepath.Join(uploadsDir, newFileName)

	err = os.WriteFile(targetPath, finalImageBytes, 0644)
	if err != nil {
		log.Printf("GeneratePhotoshoot: Failed to save file %s: %v", targetPath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save photoshoot image: " + err.Error()})
		return
	}

	url := fmt.Sprintf("/uploads/%s", newFileName)
	c.JSON(http.StatusOK, gin.H{
		"url":    url,
		"prompt": finalPrompt,
	})
}

func generatePromptWithOpenAI(base64Data string, mimeType string, name string, apiKey string) (string, error) {
	promptText := "You are a professional photographer. Look at this product image. Write a highly detailed prompt (in English) to generate a commercial-grade 8k food photoshoot of this product in a clean studio setting, aspect ratio 1:1, beautiful ambient studio lighting. Focus on product branding, style, and texture. Return ONLY a JSON object matching this schema: {\"prompt\": \"your detailed photoshoot prompt\"}"
	if name != "" {
		promptText = fmt.Sprintf("You are a professional photographer. Look at this product image of '%s'. Write a highly detailed prompt (in English) to generate a commercial-grade 8k food photoshoot of this product in a clean studio setting, aspect ratio 1:1, beautiful ambient studio lighting. Focus on product branding, style, and texture. Return ONLY a JSON object matching this schema: {\"prompt\": \"your detailed photoshoot prompt\"}", name)
	}

	imageURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)

	reqPayload := map[string]interface{}{
		"model": "gpt-4o-mini",
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": promptText,
					},
					{
						"type": "image_url",
						"image_url": map[string]string{
							"url": imageURL,
						},
					},
				},
			},
		},
		"response_format": map[string]string{
			"type": "json_object",
		},
	}

	reqBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(reqBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai chat api returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBytes, &openAIResp); err != nil || len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("failed to parse openai response: %v", err)
	}

	var parsedPrompt struct {
		Prompt string `json:"prompt"`
	}
	if err := json.Unmarshal([]byte(openAIResp.Choices[0].Message.Content), &parsedPrompt); err != nil {
		return "", fmt.Errorf("failed to parse prompt from assistant content: %v", err)
	}

	return parsedPrompt.Prompt, nil
}

func generateImageWithOpenAI(prompt string, apiKey string) (string, []byte, error) {
	// Request image from DALL-E 3 (defaults to returning a URL)
	reqPayload := map[string]interface{}{
		"model":  "dall-e-3",
		"prompt": prompt,
		"n":      1,
		"size":   "1024x1024",
	}

	reqBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return "", nil, err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/images/generations", bytes.NewBuffer(reqBytes))
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("openai images api returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var openAIResp struct {
		Data []struct {
			URL     string `json:"url"`
			B64JSON string `json:"b64_json"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBytes, &openAIResp); err != nil || len(openAIResp.Data) == 0 {
		return "", nil, fmt.Errorf("failed to parse openai images response: %v", err)
	}

	if openAIResp.Data[0].B64JSON != "" {
		decBytes, err := base64.StdEncoding.DecodeString(openAIResp.Data[0].B64JSON)
		if err != nil {
			return "", nil, fmt.Errorf("failed to decode b64_json: %v", err)
		}
		return "image/png", decBytes, nil
	}

	if openAIResp.Data[0].URL != "" {
		// Download image from URL
		imgGetResp, imgGetErr := http.Get(openAIResp.Data[0].URL)
		if imgGetErr != nil {
			return "", nil, fmt.Errorf("failed to download image from openai URL: %v", imgGetErr)
		}
		defer imgGetResp.Body.Close()

		if imgGetResp.StatusCode != http.StatusOK {
			return "", nil, fmt.Errorf("downloading image returned status %d", imgGetResp.StatusCode)
		}

		downloadedBytes, err := io.ReadAll(imgGetResp.Body)
		if err != nil {
			return "", nil, fmt.Errorf("failed to read downloaded image bytes: %v", err)
		}

		contentType := imgGetResp.Header.Get("Content-Type")
		return contentType, downloadedBytes, nil
	}

	return "", nil, fmt.Errorf("no image url or b64_json returned from openai")
}

func (h *GeminiHandler) GenerateZohoAIAssist(c *gin.Context) {
	name := c.Query("name")
	nameAr := c.Query("name_ar")

	if name == "" && nameAr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product name or Arabic name is required"})
		return
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GEMINI_API_KEY environment variable is not configured on the backend"})
		return
	}

	prompt := fmt.Sprintf(
		"You are an expert AI product content assistant for an e-commerce grocery and food service distributor in Saudi Arabia. "+
		"Your task is to correct, translate, clean up, and auto-complete product information based on the input name/title: "+
		"Input English Name: '%s', Input Arabic Name: '%s'.\n\n"+
		"Specifically, perform the following steps:\n"+
		"1. Analyze the inputs. Correct any language misplaced inputs (e.g. if the English name contains Arabic text, translate/move it correctly). "+
		"Generate a clean, high-quality, professional English name (marketing-friendly) and a matching Arabic name.\n"+
		"2. Generate a mouth-watering, premium marketing description in both English (description, 2-3 sentences) and Arabic (description_ar, 2-3 sentences) optimized for local tastes.\n"+
		"3. Extract the weight or volume in kg (as a double number/float, e.g. 6.0 or 0.5) from the titles if any weight/volume mention exists (e.g. 'كغ 6', '6 kg', '500g' = 0.5, etc.). If none exists, return 0.0.\n"+
		"4. Generate professional SEO metadata: meta_title (under 60 chars), meta_description (under 160 chars), and keywords (comma-separated list of 5-8 relevant search keywords).\n\n"+
		"You must return your output strictly in JSON format matching this schema:\n"+
		"{\n"+
		"  \"name\": \"Clean English Product Name\",\n"+
		"  \"name_ar\": \"Clean Arabic Product Name\",\n"+
		"  \"description\": \"Professional English description.\",\n"+
		"  \"description_ar\": \"Professional Arabic description.\",\n"+
		"  \"weight\": 6.0,\n"+
		"  \"meta_title\": \"SEO Meta Title\",\n"+
		"  \"meta_description\": \"SEO Meta Description\",\n"+
		"  \"keywords\": \"keyword1, keyword2, keyword3\"\n"+
		"}",
		name, nameAr,
	)

	reqPayload := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			ResponseMimeType: "application/json",
		},
	}

	reqBytes, err := json.Marshal(reqPayload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize request: " + err.Error()})
		return
	}

	apiURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=" + apiKey
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call Gemini API: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read Gemini API response: " + err.Error()})
		return
	}

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   fmt.Sprintf("Gemini API returned status %d", resp.StatusCode),
			"details": string(respBytes),
		})
		return
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(respBytes, &geminiResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Gemini API response: " + err.Error()})
		return
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Empty response candidates returned from Gemini"})
		return
	}

	rawText := geminiResp.Candidates[0].Content.Parts[0].Text

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(rawText), &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to parse generated assist result as structured JSON: " + err.Error(),
			"rawText": rawText,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
