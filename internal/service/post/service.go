package post

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
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
	ID          string `json:"id,omitempty"`
	TokenID     string `json:"tokenId,omitempty"`
	Title       string `json:"title,omitempty"`
	Body        string `json:"body,omitempty"`
	PublishedAt string `json:"publishedAt,omitempty"`
	MetadataURL string `json:"metadataUrl,omitempty"`
}

type metadata struct {
	Title       string `json:"title"`
	Body        string `json:"body"`
	PublishedAt string `json:"publishedAt"`
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

	callResults := make([]interface{}, 1)
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
			posts[i] = Post{
				ID:          tokenID,
				TokenID:     tokenID,
				Title:       meta.Title,
				Body:        meta.Body,
				PublishedAt: meta.PublishedAt,
				MetadataURL: p.Uri,
			}
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

func parsePublishedAt(val string) time.Time {
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(val))
	if err != nil {
		return time.Time{}
	}
	return t
}

func (s *Service) Create(p Post) error { return nil }
