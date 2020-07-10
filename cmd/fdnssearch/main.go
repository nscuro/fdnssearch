package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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

	pDatasetFiles  []string
	pSearchDomains []string
	pSearchTypes   []string
	pConcurrency   int
	pAny           bool
	pShowValue     bool
	pShowType      bool
	pTimeout       int64
)

func init() {
	cmd.Flags().StringArrayVarP(&pDatasetFiles, "files", "f", make([]string, 0), "dataset files")
	cmd.Flags().StringArrayVarP(&pSearchDomains, "domains", "d", make([]string, 0), "domains to search for")
	cmd.Flags().StringArrayVarP(&pSearchTypes, "types", "t", []string{"a"}, "record types to search for (a, aaaa, cname, txt, mx)")
	cmd.Flags().IntVarP(&pConcurrency, "concurrency", "c", 10, "number of concurrent search workers")
	cmd.Flags().BoolVar(&pAny, "any", false, "additionally search ANY dataset (ignored when -f is set)")
	cmd.Flags().BoolVar(&pShowValue, "show-value", false, "show record value for search results")
	cmd.Flags().BoolVar(&pShowType, "show-type", false, "show record type for search results")
	cmd.Flags().Int64Var(&pTimeout, "timeout", 0, "timeout in seconds")

	cmd.MarkFlagRequired("domains")
}

func runCmd(_ *cobra.Command, _ []string) {
	searcher := search.NewSearcher(pConcurrency)

	// TODO: Reduce redundancy...

	var ctx context.Context
	if pTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(pTimeout)*time.Second)
		defer cancel()
	} else {
		ctx = context.Background()
	}

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

			resChan, errChan, err := searcher.Search(ctx, search.Options{
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
					if (pAny && datasetType == "any") || strings.ToLower(searchType) == datasetType {
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
			fmt.Printf(aurora.Sprintf(aurora.Blue("searching in %s (%s)\n"), selectedDataset.URL, selectedDataset.Fingerprint))

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

			resChan, errChan, err := searcher.Search(ctx, search.Options{
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
