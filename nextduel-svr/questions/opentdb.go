package questions

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const openTDBEndpoint = "https://opentdb.com/api.php"

// OpenTDB fetches questions from https://opentdb.com.
type OpenTDB struct {
	Client     *http.Client
	Endpoint   string
	Difficulty string // "", "easy", "medium" or "hard"
}

func NewOpenTDB() *OpenTDB {
	return &OpenTDB{
		Client:     &http.Client{Timeout: 5 * time.Second},
		Endpoint:   openTDBEndpoint,
		Difficulty: "easy",
	}
}

type openTDBResponse struct {
	ResponseCode int               `json:"response_code"`
	Results      []openTDBQuestion `json:"results"`
}

type openTDBQuestion struct {
	Category         string   `json:"category"`
	Difficulty       string   `json:"difficulty"`
	Question         string   `json:"question"`
	CorrectAnswer    string   `json:"correct_answer"`
	IncorrectAnswers []string `json:"incorrect_answers"`
}

func (o *OpenTDB) Questions(ctx context.Context, n int) ([]Question, error) {
	params := url.Values{}
	params.Set("amount", strconv.Itoa(n))
	params.Set("type", "multiple")
	if o.Difficulty != "" {
		params.Set("difficulty", o.Difficulty)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.Endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := o.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("opentdb request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("opentdb status %d", resp.StatusCode)
	}

	var body openTDBResponse
	if err := json.NewDecoder(http.MaxBytesReader(nil, resp.Body, 1<<20)).Decode(&body); err != nil {
		return nil, fmt.Errorf("opentdb decode: %w", err)
	}
	// 0 = success; 5 = rate limited; others = bad parameters / no results.
	if body.ResponseCode != 0 {
		return nil, fmt.Errorf("opentdb response code %d", body.ResponseCode)
	}

	out := make([]Question, 0, len(body.Results))
	for _, r := range body.Results {
		// OpenTDB HTML-encodes its default output (&quot;, &#039; ...).
		text := html.UnescapeString(r.Question)
		incorrect := make([]string, len(r.IncorrectAnswers))
		for i, a := range r.IncorrectAnswers {
			incorrect[i] = html.UnescapeString(a)
		}
		out = append(out, buildQuestion(
			HashID(text), text, html.UnescapeString(r.CorrectAnswer), incorrect,
			html.UnescapeString(r.Category), r.Difficulty,
		))
	}
	return out, nil
}
