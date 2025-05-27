package anime

type Episode struct {
	TotalPages    int             `json:"total_pages"`
	TotalEpisodes int             `json:"total_episodes"`
	LastEpisode   int             `json:"last_episode"`
	Page          int             `json:"page"`
	Episodes      []LatestEpisode `json:"episodes"`
}

type LatestEpisode struct {
	Slug    string `json:"slug"`
	Img     string `json:"img"`
	Title   string `json:"title"`
	Episode string `json:"episode"`
}

type Anime struct {
	Title          string                 `json:"title"`
	Slug           string                 `json:"slug"`
	Img            string                 `json:"img"`
	Synopsis       string                 `json:"synopsis"`
	AdditionalInfo map[string]interface{} `json:"additional_info"`
}

type Server struct {
	Server string `json:"server"`
	Remote string `json:"remote"`
}
