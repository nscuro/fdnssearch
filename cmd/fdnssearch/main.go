package main

import (
	"github.com/klauspost/pgzip"
	"github.com/nscuro/fdnssearch/internal/dataset"
	"github.com/nscuro/fdnssearch/internal/search"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	cmd = &cobra.Command{
		Use: "fdnssearch",
		Run: runCmd,
	}

	pDatasetFiles    []string
	pSearchDomains   []string
	pSearchTypes     []string
	pConcurrency     int
	pAlwaysSearchAny bool
)

func init() {
	cmd.Flags().StringArrayVarP(&pDatasetFiles, "files", "f", make([]string, 0), "dataset files")
	cmd.Flags().StringArrayVarP(&pSearchDomains, "domains", "d", make([]string, 0), "domains to search for")
	cmd.Flags().StringArrayVarP(&pSearchTypes, "types", "t", []string{"a"}, "record types to search for")
	cmd.Flags().IntVarP(&pConcurrency, "concurrency", "c", 10, "number of concurrent search workers")
	cmd.Flags().BoolVarP(&pAlwaysSearchAny, "always-any", "a", false, "always search ANY dataset")
}

func runCmd(_ *cobra.Command, _ []string) {
	searcher := search.NewSearcher(pConcurrency)

	if len(pDatasetFiles) > 0 {
		for _, filePath := range pDatasetFiles {
			log.Printf("searching in %s\n", filePath)

			file, err := os.Open(filePath)
			if err != nil {
				log.Fatal(err)
				return
			}

			gzipReader, err := pgzip.NewReader(file)
			if err != nil {
				log.Fatal(err)
				return
			}

			err = searcher.Search(search.Options{
				DatasetReader: gzipReader,
				Domains:       pSearchDomains,
				Types:         pSearchTypes,
			})
			if err != nil {
				log.Fatal(err)
				return
			}

			gzipReader.Close()
			file.Close()
		}
	} else {
		datasets, err := dataset.FetchDatasets()
		if err != nil {
			log.Fatal(err)
			return
		}

		selectedDatasets := make([]dataset.Dataset, 0)
		for _, searchType := range pSearchTypes {
			for _, ds := range datasets {
				for _, datasetType := range ds.Types {
					if strings.ToLower(searchType) == datasetType {
						selectedDatasets = append(selectedDatasets, ds)
						break
					}
				}
			}
		}

		if len(selectedDatasets) == 0 {
			log.Fatalf("no matching datasets for types %v found", pSearchTypes)
			return
		}

		for _, selectedDataset := range selectedDatasets {
			log.Printf("searching in %s\n", selectedDataset.URL)

			res, err := http.Get(selectedDataset.URL)
			if err != nil {
				log.Fatal(err)
				return
			}

			gzipReader, err := pgzip.NewReader(res.Body)
			if err != nil {
				log.Fatal(err)
				return
			}

			err = searcher.Search(search.Options{
				DatasetReader: gzipReader,
				Domains:       pSearchDomains,
				Types:         pSearchTypes,
			})
			if err != nil {
				log.Fatal(err)
				return
			}

			gzipReader.Close()
			res.Body.Close()
		}
	}
}

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
