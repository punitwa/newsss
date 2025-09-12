package services

import (
	"context"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"news-aggregator/internal/models"

	"github.com/rs/zerolog"
)

// SimpleNLPClient provides basic NLP functionality without external dependencies
type SimpleNLPClient struct {
	logger zerolog.Logger
}

// NewSimpleNLPClient creates a new simple NLP client
func NewSimpleNLPClient(logger zerolog.Logger) *SimpleNLPClient {
	return &SimpleNLPClient{
		logger: logger.With().Str("component", "nlp_client").Logger(),
	}
}

// AnalyzeContent performs basic content analysis
func (c *SimpleNLPClient) AnalyzeContent(ctx context.Context, title, content string) (*models.ContentAnalysis, error) {
	c.logger.Debug().Str("title", title).Msg("Analyzing content")

	analysis := &models.ContentAnalysis{
		ProcessedAt: time.Now(),
	}

	// Calculate sentiment score (basic keyword-based approach)
	analysis.SentimentScore = c.calculateSentiment(title + " " + content)

	// Calculate importance score
	analysis.ImportanceScore = c.calculateImportance(title, content)

	// Calculate readability score
	analysis.ReadabilityScore = c.calculateReadability(content)

	// Extract keywords
	analysis.KeywordsExtracted = c.extractKeywords(title + " " + content)

	// Extract basic entities
	analysis.EntitiesExtracted = c.extractEntities(title + " " + content)

	// Classify topic
	analysis.TopicClassification = c.classifyTopic(title + " " + content)

	// Detect language (simple approach)
	analysis.LanguageDetected = c.detectLanguage(content)

	return analysis, nil
}

// ExtractKeywords extracts important keywords from text
func (c *SimpleNLPClient) ExtractKeywords(ctx context.Context, text string) ([]string, error) {
	return c.extractKeywords(text), nil
}

// ClassifyTopic classifies the topic of the text
func (c *SimpleNLPClient) ClassifyTopic(ctx context.Context, text string) (string, error) {
	return c.classifyTopic(text), nil
}

// CalculateImportance calculates the importance score of the content
func (c *SimpleNLPClient) CalculateImportance(ctx context.Context, title, content string) (float64, error) {
	return c.calculateImportance(title, content), nil
}

// calculateSentiment performs basic sentiment analysis
func (c *SimpleNLPClient) calculateSentiment(text string) float64 {
	text = strings.ToLower(text)

	// Positive words
	positiveWords := []string{
		"good", "great", "excellent", "amazing", "wonderful", "fantastic", "awesome",
		"positive", "success", "win", "victory", "achievement", "breakthrough", "progress",
		"improve", "better", "best", "outstanding", "remarkable", "impressive", "brilliant",
		"celebrate", "happy", "joy", "pleased", "satisfied", "delighted", "thrilled",
	}

	// Negative words
	negativeWords := []string{
		"bad", "terrible", "awful", "horrible", "disaster", "crisis", "problem", "issue",
		"negative", "fail", "failure", "loss", "defeat", "decline", "drop", "fall",
		"worse", "worst", "concerning", "worried", "alarming", "dangerous", "threat",
		"sad", "angry", "upset", "disappointed", "frustrated", "concerned", "fear",
	}

	positiveCount := 0
	negativeCount := 0

	words := strings.Fields(text)
	for _, word := range words {
		word = strings.Trim(word, ".,!?;:")
		for _, positive := range positiveWords {
			if word == positive {
				positiveCount++
				break
			}
		}
		for _, negative := range negativeWords {
			if word == negative {
				negativeCount++
				break
			}
		}
	}

	totalSentimentWords := positiveCount + negativeCount
	if totalSentimentWords == 0 {
		return 0.0 // Neutral
	}

	// Return score between -1.0 and 1.0
	return float64(positiveCount-negativeCount) / float64(totalSentimentWords)
}

// calculateImportance calculates content importance based on various factors
func (c *SimpleNLPClient) calculateImportance(title, content string) float64 {
	score := 0.5 // Base score

	// Title factors
	titleWords := strings.Fields(strings.ToLower(title))

	// Important keywords in title
	importantKeywords := []string{
		"breaking", "urgent", "major", "significant", "important", "critical",
		"exclusive", "first", "new", "latest", "update", "announced",
		"government", "president", "minister", "election", "policy",
		"economy", "market", "stock", "financial", "business",
		"technology", "ai", "innovation", "research", "study",
		"health", "medical", "pandemic", "vaccine", "treatment",
		"climate", "environment", "global", "international", "world",
	}

	for _, word := range titleWords {
		for _, keyword := range importantKeywords {
			if word == keyword {
				score += 0.05
				break
			}
		}
	}

	// Content length factor
	contentLength := len(content)
	if contentLength >= 500 && contentLength <= 3000 {
		score += 0.1
	} else if contentLength > 3000 {
		score += 0.05
	}

	// Numbers and statistics (often indicate factual content)
	numberRegex := regexp.MustCompile(`\d+`)
	numbers := numberRegex.FindAllString(content, -1)
	if len(numbers) > 5 {
		score += 0.1
	}

	// Quotes (indicate interviews or official statements)
	quoteCount := strings.Count(content, "\"")
	if quoteCount >= 4 {
		score += 0.05
	}

	// Capitalize entities (proper nouns)
	capitalizedWords := regexp.MustCompile(`\b[A-Z][a-z]+\b`).FindAllString(content, -1)
	if len(capitalizedWords) > 10 {
		score += 0.05
	}

	return math.Min(score, 1.0)
}

// calculateReadability calculates basic readability score
func (c *SimpleNLPClient) calculateReadability(content string) float64 {
	if len(content) == 0 {
		return 0.0
	}

	sentences := strings.Split(content, ".")
	words := strings.Fields(content)

	if len(sentences) == 0 || len(words) == 0 {
		return 0.5
	}

	// Average words per sentence
	avgWordsPerSentence := float64(len(words)) / float64(len(sentences))

	// Count syllables (simplified)
	totalSyllables := 0
	for _, word := range words {
		totalSyllables += c.countSyllables(word)
	}
	avgSyllablesPerWord := float64(totalSyllables) / float64(len(words))

	// Simplified Flesch Reading Ease formula
	// Score = 206.835 - (1.015 × ASL) - (84.6 × ASW)
	score := 206.835 - (1.015 * avgWordsPerSentence) - (84.6 * avgSyllablesPerWord)

	// Normalize to 0-1 range (typical scores range from 0-100)
	normalizedScore := math.Max(0, math.Min(100, score)) / 100.0

	return normalizedScore
}

// countSyllables counts syllables in a word (simplified approach)
func (c *SimpleNLPClient) countSyllables(word string) int {
	word = strings.ToLower(word)
	vowels := "aeiouy"
	syllables := 0
	prevWasVowel := false

	for _, char := range word {
		isVowel := strings.ContainsRune(vowels, char)
		if isVowel && !prevWasVowel {
			syllables++
		}
		prevWasVowel = isVowel
	}

	// Handle silent 'e'
	if strings.HasSuffix(word, "e") && syllables > 1 {
		syllables--
	}

	// Minimum of 1 syllable per word
	if syllables == 0 {
		syllables = 1
	}

	return syllables
}

// extractKeywords extracts important keywords from text
func (c *SimpleNLPClient) extractKeywords(text string) []string {
	text = strings.ToLower(text)

	// Remove common stop words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "from": true, "up": true, "about": true, "into": true,
		"through": true, "during": true, "before": true, "after": true, "above": true,
		"below": true, "between": true, "among": true, "is": true, "are": true, "was": true,
		"were": true, "be": true, "been": true, "being": true, "have": true, "has": true,
		"had": true, "do": true, "does": true, "did": true, "will": true, "would": true,
		"could": true, "should": true, "may": true, "might": true, "must": true,
		"this": true, "that": true, "these": true, "those": true, "i": true, "me": true,
		"my": true, "myself": true, "we": true, "our": true, "ours": true, "ourselves": true,
		"you": true, "your": true, "yours": true, "yourself": true, "yourselves": true,
		"he": true, "him": true, "his": true, "himself": true, "she": true, "her": true,
		"hers": true, "herself": true, "it": true, "its": true, "itself": true,
		"they": true, "them": true, "their": true, "theirs": true, "themselves": true,
	}

	// Extract words
	wordRegex := regexp.MustCompile(`\b[a-zA-Z]{3,}\b`)
	words := wordRegex.FindAllString(text, -1)

	// Count word frequency
	wordCount := make(map[string]int)
	for _, word := range words {
		word = strings.ToLower(word)
		if !stopWords[word] {
			wordCount[word]++
		}
	}

	// Sort by frequency
	type wordFreq struct {
		word  string
		count int
	}

	var wordFreqs []wordFreq
	for word, count := range wordCount {
		if count >= 2 { // Only include words that appear at least twice
			wordFreqs = append(wordFreqs, wordFreq{word, count})
		}
	}

	sort.Slice(wordFreqs, func(i, j int) bool {
		return wordFreqs[i].count > wordFreqs[j].count
	})

	// Return top keywords
	var keywords []string
	maxKeywords := 10
	for i, wf := range wordFreqs {
		if i >= maxKeywords {
			break
		}
		keywords = append(keywords, wf.word)
	}

	return keywords
}

// extractEntities extracts basic named entities
func (c *SimpleNLPClient) extractEntities(text string) map[string]string {
	entities := make(map[string]string)

	// Simple patterns for entity extraction
	patterns := map[string]*regexp.Regexp{
		"PERSON":       regexp.MustCompile(`\b[A-Z][a-z]+ [A-Z][a-z]+\b`),
		"ORGANIZATION": regexp.MustCompile(`\b[A-Z][A-Z]+\b|\b[A-Z][a-z]+ [A-Z][a-z]+\b(?:\s+(?:Inc|Corp|Ltd|LLC|Company|Organization|University|College))?`),
		"DATE":         regexp.MustCompile(`\b(?:January|February|March|April|May|June|July|August|September|October|November|December)\s+\d{1,2},?\s+\d{4}\b|\b\d{1,2}/\d{1,2}/\d{4}\b`),
		"MONEY":        regexp.MustCompile(`\$\d+(?:,\d{3})*(?:\.\d{2})?|\b\d+(?:,\d{3})*(?:\.\d{2})?\s+(?:dollars?|USD|euros?|EUR|pounds?|GBP)\b`),
		"PERCENTAGE":   regexp.MustCompile(`\b\d+(?:\.\d+)?%\b`),
	}

	for entityType, pattern := range patterns {
		matches := pattern.FindAllString(text, -1)
		for _, match := range matches {
			entities[strings.TrimSpace(match)] = entityType
		}
	}

	return entities
}

// classifyTopic classifies the main topic of the text
func (c *SimpleNLPClient) classifyTopic(text string) string {
	text = strings.ToLower(text)

	topicKeywords := map[string][]string{
		"technology":    {"technology", "tech", "ai", "artificial", "intelligence", "computer", "software", "hardware", "internet", "digital", "cyber", "innovation", "startup", "app", "platform"},
		"business":      {"business", "economy", "economic", "market", "stock", "financial", "finance", "investment", "company", "corporate", "trade", "industry", "revenue", "profit", "sales"},
		"politics":      {"politics", "political", "government", "election", "vote", "president", "minister", "congress", "parliament", "policy", "law", "legislation", "campaign", "democracy"},
		"health":        {"health", "medical", "medicine", "doctor", "hospital", "patient", "treatment", "disease", "vaccine", "pandemic", "virus", "healthcare", "wellness", "fitness"},
		"sports":        {"sports", "sport", "game", "match", "team", "player", "football", "basketball", "cricket", "tennis", "soccer", "baseball", "championship", "tournament", "league"},
		"science":       {"science", "scientific", "research", "study", "experiment", "discovery", "climate", "environment", "space", "nasa", "physics", "chemistry", "biology", "genetics"},
		"entertainment": {"entertainment", "movie", "film", "actor", "actress", "music", "singer", "celebrity", "hollywood", "tv", "television", "show", "concert", "album", "award"},
		"world":         {"world", "international", "global", "country", "nation", "war", "conflict", "peace", "diplomacy", "foreign", "embassy", "united nations", "europe", "asia", "africa"},
	}

	topicScores := make(map[string]int)
	words := strings.Fields(text)

	for _, word := range words {
		word = strings.Trim(word, ".,!?;:")
		for topic, keywords := range topicKeywords {
			for _, keyword := range keywords {
				if strings.Contains(word, keyword) || strings.Contains(keyword, word) {
					topicScores[topic]++
				}
			}
		}
	}

	// Find topic with highest score
	maxScore := 0
	bestTopic := "general"
	for topic, score := range topicScores {
		if score > maxScore {
			maxScore = score
			bestTopic = topic
		}
	}

	return bestTopic
}

// detectLanguage performs basic language detection
func (c *SimpleNLPClient) detectLanguage(text string) string {
	// Very basic language detection based on common words
	text = strings.ToLower(text)

	englishWords := []string{"the", "and", "of", "to", "a", "in", "is", "it", "you", "that", "he", "was", "for", "on", "are", "as", "with", "his", "they", "i"}
	spanishWords := []string{"el", "la", "de", "que", "y", "a", "en", "un", "es", "se", "no", "te", "lo", "le", "da", "su", "por", "son", "con", "para"}
	frenchWords := []string{"le", "de", "et", "à", "un", "il", "être", "et", "en", "avoir", "que", "pour", "dans", "ce", "son", "une", "sur", "avec", "ne", "se"}

	englishCount := 0
	spanishCount := 0
	frenchCount := 0

	words := strings.Fields(text)
	for _, word := range words {
		word = strings.Trim(word, ".,!?;:")

		for _, engWord := range englishWords {
			if word == engWord {
				englishCount++
				break
			}
		}

		for _, spaWord := range spanishWords {
			if word == spaWord {
				spanishCount++
				break
			}
		}

		for _, freWord := range frenchWords {
			if word == freWord {
				frenchCount++
				break
			}
		}
	}

	// Return language with highest count
	if englishCount >= spanishCount && englishCount >= frenchCount {
		return "en"
	} else if spanishCount >= frenchCount {
		return "es"
	} else if frenchCount > 0 {
		return "fr"
	}

	return "en" // Default to English
}
