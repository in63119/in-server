package post

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"golang.org/x/sync/errgroup"

	"in-server/pkg/apperr"
	"in-server/pkg/config"
	"in-server/pkg/eth"
	"in-server/pkg/types"
)

type Service struct {
	cfg        config.Config
	eth        *eth.Client
	httpClient *http.Client
}

func New(ctx context.Context, cfg config.Config) (*Service, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	ethClient, err := eth.Dial(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("dial eth client: %w", err)
	}

	return &Service{
		cfg:        cfg,
		eth:        ethClient,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

type Post struct {
	ID                 string   `json:"id,omitempty"`
	TokenID            string   `json:"tokenId,omitempty"`
	Slug               string   `json:"slug,omitempty"`
	Title              string   `json:"title,omitempty"`
	Summary            string   `json:"summary,omitempty"`
	Description        string   `json:"description,omitempty"`
	Category           string   `json:"category,omitempty"`
	LabName            string   `json:"labName,omitempty"`
	LabSegment         string   `json:"labSegment,omitempty"`
	Href               string   `json:"href,omitempty"`
	PublishedAt        string   `json:"publishedAt,omitempty"`
	ReadingTimeMinutes int      `json:"readingTimeMinutes,omitempty"`
	ReadingTimeLabel   string   `json:"readingTimeLabel,omitempty"`
	Tags               []string `json:"tags,omitempty"`
	MetadataURL        string   `json:"metadataUrl,omitempty"`
	Image              string   `json:"image,omitempty"`
	ExternalURL        string   `json:"externalUrl,omitempty"`
	Content            string   `json:"content,omitempty"`
	RelatedLinks       []string `json:"relatedLinks,omitempty"`
	StructuredData     string   `json:"structuredData,omitempty"`
}

type metadata struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Image       string      `json:"image"`
	ExternalURL string      `json:"external_url"`
	Attributes  []attribute `json:"attributes"`
}

type attribute struct {
	TraitType   string          `json:"trait_type"`
	Value       json.RawMessage `json:"value"`
	DisplayType string          `json:"display_type,omitempty"`
}

func (s *Service) List() ([]Post, error) {
	ctx := context.Background()

	if s.eth == nil {
		return nil, fmt.Errorf("eth client is nil")
	}

	adminCode := strings.TrimSpace(s.cfg.Auth.AdminCode)
	if adminCode == "" {
		return nil, apperr.System.ErrMissingAuthAdminCode
	}

	_, ownerAddr, err := s.eth.Wallet(adminCode)
	if err != nil {
		return nil, fmt.Errorf("owner address: %w", err)
	}

	contract, _, err := s.eth.Contract(types.POSTSTORAGE)
	if err != nil {
		return nil, fmt.Errorf("bind post storage: %w", err)
	}

	var raw []struct {
		Id  *big.Int
		Uri string
	}

	callOpts := &bind.CallOpts{Context: ctx}

	callResults := make([]any, 1)
	callResults[0] = new([]struct {
		Id  *big.Int
		Uri string
	})

	if err := contract.Call(callOpts, &callResults, "getPosts", ownerAddr); err != nil {
		return nil, fmt.Errorf("call getPosts: %w", err)
	}

	rawPtr, ok := callResults[0].(*[]struct {
		Id  *big.Int
		Uri string
	})
	if !ok || rawPtr == nil {
		return nil, fmt.Errorf("unexpected call result type for getPosts")
	}
	raw = *rawPtr

	posts := make([]Post, len(raw))
	eg, egctx := errgroup.WithContext(ctx)
	for i, p := range raw {
		i, p := i, p
		eg.Go(func() error {
			meta, err := s.fetchMetadata(egctx, p.Uri)
			if err != nil {
				return fmt.Errorf("fetch metadata for %s: %w", p.Uri, err)
			}
			tokenID := p.Id.String()
			posts[i] = mapMetadataToPost(meta, tokenID, p.Uri)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	sort.Slice(posts, func(i, j int) bool {
		return parsePublishedAt(posts[i].PublishedAt).After(parsePublishedAt(posts[j].PublishedAt))
	})

	return posts, nil
}

func (s *Service) fetchMetadata(ctx context.Context, url string) (metadata, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return metadata{}, apperr.System.ErrInvalidMetadataURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return metadata{}, fmt.Errorf("build request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return metadata{}, apperr.Wrap(err, apperr.System.ErrInvalidMetadataURL.Code, "http get metadata", apperr.System.ErrLoadMetadata.Status)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return metadata{}, apperr.Wrap(fmt.Errorf("status %d", resp.StatusCode), apperr.System.ErrLoadMetadata.Code, "unexpected metadata status", apperr.System.ErrLoadMetadata.Status)
	}

	var meta metadata
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return metadata{}, fmt.Errorf("decode metadata: %w", err)
	}
	return meta, nil
}

func mapMetadataToPost(meta metadata, tokenID, metadataURL string) Post {
	var slug, summary, category, labName, labSegment, href, publishedAt, readingLabel, content, structuredData string
	var readingMinutes int
	var tags []string
	var related []string

	title := strings.TrimSpace(meta.Name)
	description := strings.TrimSpace(meta.Description)
	image := strings.TrimSpace(meta.Image)
	externalURL := strings.TrimSpace(meta.ExternalURL)

	for _, attr := range meta.Attributes {
		key := strings.TrimSpace(attr.TraitType)
		val := strings.TrimSpace(attrString(attr.Value))

		switch key {
		case "Slug":
			slug = val
		case "Summary":
			summary = val
		case "Content":
			content = val
		case "PublishedAt":
			publishedAt = val
		case "ReadingTimeMinutes":
			if n := attrInt(attr.Value); n > 0 {
				readingMinutes = n
				readingLabel = fmt.Sprintf("%d min read", n)
			}
		case "Tags":
			if val != "" {
				tags = strings.Fields(val)
			}
		case "Lab":
			labName = val
		case "StructuredData":
			structuredData = val
		}
	}

	if labSegment == "" {
		if seg := firstPathSegment(externalURL); seg != "" {
			labSegment = seg
		}
	}
	if category == "" && strings.Contains(labSegment, "-") {
		category = strings.SplitN(labSegment, "-", 2)[0]
	}
	if href == "" && externalURL != "" {
		href = pathFromURL(externalURL)
	}

	return Post{
		ID:                 tokenID,
		TokenID:            tokenID,
		Slug:               slug,
		Title:              title,
		Summary:            summary,
		Description:        description,
		Category:           category,
		LabName:            labName,
		LabSegment:         labSegment,
		Href:               href,
		PublishedAt:        publishedAt,
		ReadingTimeMinutes: readingMinutes,
		ReadingTimeLabel:   readingLabel,
		Tags:               tags,
		MetadataURL:        metadataURL,
		Image:              image,
		ExternalURL:        externalURL,
		Content:            content,
		RelatedLinks:       related,
		StructuredData:     structuredData,
	}
}

func attrString(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.6f", f), "0"), ".")
	}
	return string(raw)
}

func attrInt(raw json.RawMessage) int {
	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		return n
	}
	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		return int(f)
	}
	return 0
}

func firstPathSegment(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func pathFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	if u.Path == "" {
		return rawURL
	}
	return u.Path
}

func parsePublishedAt(val string) time.Time {
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(val))
	if err != nil {
		return time.Time{}
	}
	return t
}

func (s *Service) Create(p Post) error { return nil }
