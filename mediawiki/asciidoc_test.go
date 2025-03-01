package mediawiki

import (
	"strings"
	"testing"
)

func TestParseWikitext(t *testing.T) {
	tests := []struct {
		name     string
		input    PageResult
		expected string
		wantErr  bool
	}{
		{
			name: "description list",
			input: PageResult{
				Parse: struct {
					Title    string `json:"title"`
					Wikitext struct {
						All string `json:"*"`
					} `json:"wikitext"`
				}{
					Wikitext: struct {
						All string `json:"*"`
					}{
						All: "\n\n:: Ellora Section: [[Serapis Bey]] \n:: Section of Solomon: [[Polydorus Isurenus]] \n:: Section of the Serpent: [[The Serpent]]\n\n",
					},
				},
			},
			expected: "\n\nEllora Section:: [[Serapis Bey]] \n\nSection of Solomon:: [[Polydorus Isurenus]] \n\nSection of the Serpent:: [[The Serpent]]\n\n",
			wantErr:  false,
		},
		{
			name: "description list with double newlines",
			input: PageResult{
				Parse: struct {
					Title    string `json:"title"`
					Wikitext struct {
						All string `json:"*"`
					} `json:"wikitext"`
				}{
					Wikitext: struct {
						All string `json:"*"`
					}{
						All: "\n\n:: Ellora Section: [[Serapis Bey]] \n\n:: Section of Solomon: [[Polydorus Isurenus]] \n\n:: Section of the Serpent: [[The Serpent]]\n\n",
					},
				},
			},
			expected: "\n\nEllora Section:: [[Serapis Bey]] \n\nSection of Solomon:: [[Polydorus Isurenus]] \n\nSection of the Serpent:: [[The Serpent]]\n\n",
			wantErr:  false,
		},
		{
			name: "wikilink conversion",
			input: PageResult{
				Parse: struct {
					Title    string `json:"title"`
					Wikitext struct {
						All string `json:"*"`
					} `json:"wikitext"`
				}{
					Wikitext: struct {
						All string `json:"*"`
					}{
						All: "[[Helena Petrovna Blavatsky|H. P. Blavatsky]]'s writing room at [[Adyar (campus)|Adyar]] (not fixed to it).",
					},
				},
			},
			expected: "[[Helena Petrovna Blavatsky|H. P. Blavatsky]]'s writing room at [[Adyar (campus)|Adyar]] (not fixed to it).",
			wantErr:  false,
		},
		{
			name: "new line after section header",
			input: PageResult{
				Parse: struct {
					Title    string `json:"title"`
					Wikitext struct {
						All string `json:"*"`
					} `json:"wikitext"`
				}{
					Wikitext: struct {
						All string `json:"*"`
					}{
						All: "== China tray phenomenon ==\n\nThe following phenomena, stated by",
					},
				},
			},
			expected: "=== China tray phenomenon\n\nThe following phenomena, stated by",
			wantErr:  false,
		},
		{
			name: "section header is not the first line",
			input: PageResult{
				Parse: struct {
					Title    string `json:"title"`
					Wikitext struct {
						All string `json:"*"`
					} `json:"wikitext"`
				}{
					Wikitext: struct {
						All string `json:"*"`
					}{
						All: "\n\n== Early life and education ==\n\nA. Trevor Barker was born at Las Palmas in the Canary Islands, on [[October 10]], 1893.",
					},
				},
			},
			expected: "=== Early life and education\n\nA. Trevor Barker was born at Las Palmas in the Canary Islands, on [[October 10]], 1893.",
			wantErr:  false,
		},
		{
			name: "footnotes",
			input: PageResult{
				Parse: struct {
					Title    string `json:"title"`
					Wikitext struct {
						All string `json:"*"`
					} `json:"wikitext"`
				}{
					Wikitext: struct {
						All string `json:"*"`
					}{
						All: "This Brotherhood has several Sections, as can be seen in one of the letters [[Master]] [[Tuitit Bey]] sent to [[H. S. Olcott]]:\u003Cref\u003ECuruppumullage Jinarajadasa, ''Letters from the Masters of the Wisdom'' Second Series, Letter No. 3 (Adyar, Madras: Theosophical Publishing House, 1977), 18. In 1926 edition, see page 21.\u003C/ref\u003E",
					},
				},
			},
			expected: "This Brotherhood has several Sections, as can be seen in one of the letters [[Master]] [[Tuitit Bey]] sent to [[H. S. Olcott]]:footnote:[Curuppumullage Jinarajadasa, _Letters from the Masters of the Wisdom_ Second Series, Letter No. 3 (Adyar, Madras: Theosophical Publishing House, 1977), 18. In 1926 edition, see page 21.]",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseWikitext(tt.input, "lua")
			if (err != nil) != tt.wantErr {
				t.Errorf("parseWikitext() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			// Since pandoc might add/modify whitespace, we'll compare trimmed strings
			got = strings.TrimSpace(got)
			expected := strings.TrimSpace(tt.expected)

			if !strings.Contains(got, expected) {
				t.Errorf("parseWikitext() = \n%v\n<<WANT>>\n%v", got, expected)
			}
		})
	}
}
