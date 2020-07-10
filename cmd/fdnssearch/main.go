package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/klauspost/pgzip"
	"github.com/logrusorgru/aurora"
	"github.com/nscuro/fdnssearch/internal/dataset"
	"github.com/nscuro/fdnssearch/internal/search"
	"github.com/spf13/cobra"
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
	pShowValue       bool
	pShowType        bool
)

func init() {
	cmd.Flags().StringArrayVarP(&pDatasetFiles, "files", "f", make([]string, 0), "dataset files")
	cmd.Flags().StringArrayVarP(&pSearchDomains, "domains", "d", make([]string, 0), "domains to search for")
	cmd.Flags().StringArrayVarP(&pSearchTypes, "types", "t", []string{"a"}, "record types to search for")
	cmd.Flags().IntVarP(&pConcurrency, "concurrency", "c", 10, "number of concurrent search workers")
	cmd.Flags().BoolVar(&pAlwaysSearchAny, "always-any", false, "always search ANY dataset (ignored when -f is set)")
	cmd.Flags().BoolVar(&pShowValue, "show-value", false, "show record value for search results")
	cmd.Flags().BoolVar(&pShowType, "show-type", false, "show record type for search results")
}

func runCmd(_ *cobra.Command, _ []string) {
	searcher := search.NewSearcher(pConcurrency)

	// TODO: Reduce redundancy...

	if len(pDatasetFiles) > 0 {
		for _, filePath := range pDatasetFiles {
			fmt.Printf(aurora.Sprintf(aurora.Blue("searching in %s\n"), filePath))

			file, err := os.Open(filePath)
			if err != nil {
				fmt.Printf(aurora.Sprintf(aurora.Red("%v\n"), err))
				return
			}

			gzipReader, err := pgzip.NewReader(file)
			if err != nil {
				fmt.Printf(aurora.Sprintf(aurora.Red("%v\n"), err))
				return
			}

			resChan, errChan, err := searcher.Search(search.Options{
				DatasetReader: gzipReader,
				Domains:       pSearchDomains,
				Types:         pSearchTypes,
			})
			if err != nil {
				fmt.Printf(aurora.Sprintf(aurora.Red("%v\n"), err))
				return
			}

			go func() {
				for err := range errChan {
					fmt.Printf(aurora.Sprintf(aurora.Red("%v\n"), err))
				}
			}()

			for res := range resChan {
				if pShowValue && pShowType {
					fmt.Printf(aurora.Sprintf(aurora.Green("%s %s %s\n"), res.Name, res.Value, res.Type))
				} else if pShowValue {
					fmt.Printf(aurora.Sprintf(aurora.Green("%s %s\n"), res.Name, res.Value))
				} else if pShowType {
					fmt.Printf(aurora.Sprintf(aurora.Green("%s %s\n"), res.Name, res.Type))
				} else {
					fmt.Printf(aurora.Sprintf(aurora.Green("%s\n"), res.Name))
				}
			}

			gzipReader.Close()
			file.Close()
		}
	} else {
		datasets, err := dataset.FetchDatasets()
		if err != nil {
			fmt.Printf(aurora.Sprintf(aurora.Red("%v\n"), err))
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
			fmt.Printf(aurora.Sprintf(aurora.Red("no matching datasets for types %v found\n"), pSearchTypes))
			return
		}

		for _, selectedDataset := range selectedDatasets {
			fmt.Printf(aurora.Sprintf(aurora.Blue("searching in %s\n"), selectedDataset.URL))

			res, err := http.Get(selectedDataset.URL)
			if err != nil {
				fmt.Printf(aurora.Sprintf(aurora.Red("%v\n"), err))
				return
			}

			gzipReader, err := pgzip.NewReader(res.Body)
			if err != nil {
				fmt.Printf(aurora.Sprintf(aurora.Red("%v\n"), err))
				return
			}

			resChan, errChan, err := searcher.Search(search.Options{
				DatasetReader: gzipReader,
				Domains:       pSearchDomains,
				Types:         pSearchTypes,
			})
			if err != nil {
				fmt.Printf(aurora.Sprintf(aurora.Red("%v\n"), err))
				return
			}

			go func() {
				for err := range errChan {
					fmt.Printf(aurora.Sprintf(aurora.Red("%v\n"), err))
				}
			}()

			for res := range resChan {
				if pShowValue && pShowType {
					fmt.Printf(aurora.Sprintf(aurora.Green("%s %s %s\n"), res.Name, res.Value, res.Type))
				} else if pShowValue {
					fmt.Printf(aurora.Sprintf(aurora.Green("%s %s\n"), res.Name, res.Value))
				} else if pShowType {
					fmt.Printf(aurora.Sprintf(aurora.Green("%s %s\n"), res.Name, res.Type))
				} else {
					fmt.Printf(aurora.Sprintf(aurora.Green("%s\n"), res.Name))
				}
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
