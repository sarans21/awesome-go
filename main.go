package main

import (
	"context"
	"log"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/google/go-github/v51/github"
	"golang.org/x/oauth2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Data struct {
	UpdatedAt string
	Nbsps     []string
	Sections  []Section
}

type Section struct {
	Title string
	Repos []Repo
}

type Repo struct {
	Org         string
	Name        string
	FullName    string
	Description string
	Stars       string
	Link        string
}

var readmeTmpl = `
# Awesome-Go

**Last update:** {{ .UpdatedAt }}

A list of my personally frequently used modules.
{{ range .Sections }}
## {{ .Title }}
|Repo{{ index $.Nbsps 0 }}|Description{{ index $.Nbsps 1 }}|Stars{{ index $.Nbsps 2 }}|
|---|---|---|{{ range .Repos }}
|[{{ .FullName }}]({{ .Link }})|{{ .Description }}|{{ .Stars }}|{{ end }}
{{ end }}
## LICENSE
MIT Saran Siriphantnon
`

var sections = []Section{
	{
		Title: "Configurations",
		Repos: []Repo{
			{Org: "kelseyhightower", Name: "envconfig"},
		},
	},
	{
		Title: "Data Types",
		Repos: []Repo{
			{Org: "shopspring", Name: "decimal"},
		},
	},
	{
		Title: "Database Clients & Tools",
		Repos: []Repo{
			{Org: "jackc", Name: "pgx"},
			{Org: "kyleconroy", Name: "sqlc"},
		},
	},
	{
		Title: "Email",
		Repos: []Repo{
			{Org: "jordan-wright", Name: "email"},
		},
	},
	{
		Title: "HTTP Clients",
		Repos: []Repo{
			{Org: "go-resty", Name: "resty"},
		},
	},
	{
		Title: "HTTP Servers",
		Repos: []Repo{
			{Org: "gofiber", Name: "fiber"},
		},
	},
	{
		Title: "Logging",
		Repos: []Repo{
			{Org: "uber-go", Name: "zap"},
		},
	},
	{
		Title: "Messaging",
		Repos: []Repo{
			{Org: "rabbitmq", Name: "amqp091-go"},
		},
	},
	{
		Title: "Testing",
		Repos: []Repo{
			{Org: "ory", Name: "dockertest"},
		},
	},
	{
		Title: "Tools",
		Repos: []Repo{
			{Org: "golangci", Name: "golangci-lint"},
			{Org: "securego", Name: "gosec"},
		},
	},
	{
		Title: "Utilities",
		Repos: []Repo{
			{Org: "mitchellh", Name: "mapstructure"},
		},
	},
	{
		Title: "Resources",
		Repos: []Repo{
			{Org: "tmrts", Name: "go-patterns"},
		},
	},
	{
		Title: "Popular",
		Repos: []Repo{},
	},
}

func main() {
	log.Println("generating README.md")
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GH_TOKEN")},
	)
	httpc := oauth2.NewClient(context.Background(), ts)
	httpc.Timeout = 2 * time.Minute
	gh := github.NewClient(httpc)
	t := template.Must(template.New("readme").Parse(readmeTmpl))
	d := Data{
		UpdatedAt: time.Now().Format("2006 Jan 2"),
		Nbsps: []string{
			strings.Repeat("&nbsp;", 40),
			strings.Repeat("&nbsp;", 90),
			strings.Repeat("&nbsp;", 7),
		},
	}
	p := message.NewPrinter(language.English)

	log.Println(p.Sprintf("%d", 5))

	for _, s := range sections {
		for i, r := range s.Repos {
			repo, _, err := gh.Repositories.Get(context.Background(), r.Org, r.Name)
			if err != nil {
				panic(err)
			}
			s.Repos[i].FullName = repo.GetFullName()
			s.Repos[i].Description = repo.GetDescription()
			s.Repos[i].Stars = p.Sprintf("%d", repo.GetStargazersCount())
			s.Repos[i].Link = repo.GetHTMLURL()
		}
	}

	log.Println("queried repos")
	d.Sections = sections

	log.Println("searching and sorting repos...")
	result, _, err := gh.Search.Repositories(context.Background(), "language: golang", &github.SearchOptions{
		Sort: "stars",
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	})
	if err != nil {
		panic(err)
	}
	for _, r := range result.Repositories {
		d.Sections[len(d.Sections)-1].Repos = append(d.Sections[len(d.Sections)-1].Repos, Repo{
			FullName:    *r.FullName,
			Description: strings.ReplaceAll(*r.Description, "|", " "),
			Stars:       p.Sprintf("%d", *r.StargazersCount),
			Link:        r.GetHTMLURL(),
		})
	}
	log.Printf("searched: %d", len(result.Repositories))

	if err := t.Execute(os.Stdout, d); err != nil {
		panic(err)
	}

	log.Println("done")
}
