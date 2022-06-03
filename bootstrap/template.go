package bootstrap

import (
	"BitSearch/searcher/words"
	"embed"
)

func SetupTemplate(tmplFS embed.FS) {
	words.DictionaryFS = tmplFS
}
