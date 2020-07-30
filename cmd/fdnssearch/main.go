package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/klauspost/pgzip"
	"github.com/nscuro/fdnssearch/internal/dataset"
	"github.com/nscuro/fdnssearch/internal/interop"
	"github.com/nscuro/fdnssearch/internal/logging"
	"github.com/nscuro/fdnssearch/internal/search"
	"github.com/spf13/cobra"
)

var (
	cmd = &cobra.Command{
		Use: "fdnssearch",
		Run: runCmd,
	}

	pgzipBlocks    = runtime.NumCPU()
	pgzipBlockSize = 5000000 // 5MB

	pDatasetFiles    []string
	pSearchDomains   []string
	pExcludedDomains []string
	pSearchTypes     []string
	pAny             bool
	pAnyOnly         bool
	pTimeout         int64
	pQuiet           bool
	pPlain           bool
	pAmassConfig     string
	pOutput          string
)

func init() {
	cmd.Flags().StringArrayVarP(&pDatasetFiles, "files", "f", make([]string, 0), "dataset files")
	cmd.Flags().StringArrayVarP(&pSearchDomains, "domains", "d", make([]string, 0), "domains to search for")
	cmd.Flags().StringArrayVarP(&pExcludedDomains, "excludes", "e", make([]string, 0), "domains to exclude from search")
	cmd.Flags().StringArrayVarP(&pSearchTypes, "types", "t", []string{"a"}, "record types to search for (a, aaaa, cname, txt, mx)")
	cmd.Flags().BoolVarP(&pAny, "any", "a", false, "additionally search ANY dataset (ignored when -f is set)")
	cmd.Flags().BoolVar(&pAnyOnly, "any-only", false, "only search ANY dataset (ignored when -f is set)")
	cmd.Flags().Int64Var(&pTimeout, "timeout", 0, "timeout in seconds")
	cmd.Flags().BoolVarP(&pQuiet, "quiet", "q", false, "only print results, no errors or log messages")
	cmd.Flags().BoolVar(&pPlain, "plain", false, "disable colored output")
	cmd.Flags().StringVar(&pAmassConfig, "amass-config", "", "amass config to load domains from")
	cmd.Flags().StringVarP(&pOutput, "output", "o", "", "output file")
}

func runCmd(_ *cobra.Command, _ []string) {
	logger := logging.NewLogger(os.Stderr, logging.Options{
		Silent:       pQuiet,
		Colorized:    !pPlain,
		ResultWriter: os.Stdout,
	})

	searcher := search.NewSearcher()

	// TODO: Reduce redundancy...

	var ctx context.Context
	if pTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(pTimeout)*time.Second)
		defer cancel()
	} else {
		ctx = context.Background()
	}

	if pAmassConfig != "" {
		logger.Infof("parsing amass config from %s", pAmassConfig)
		amassCfg, err := interop.ParseAmassConfig(pAmassConfig)
		if err != nil {
			logger.Err(err)
			return
		}
		if len(amassCfg.Domains) > 0 {
			logger.Infof("parsed domains: %v", amassCfg.Domains)
			pSearchDomains = append(pSearchDomains, amassCfg.Domains...)
		}
		if len(amassCfg.Blacklisted) > 0 {
			logger.Infof("parsed exclusions: %v", amassCfg.Blacklisted)
			pExcludedDomains = append(pExcludedDomains, amassCfg.Blacklisted...)
		}
	}

	var resultsWriter io.Writer
	if pOutput == "" || pOutput == "-" {
		resultsWriter = ioutil.Discard
	} else {
		outputFile, err := os.Create(pOutput)
		if err != nil {
			logger.Errorf("failed to create output file: %v", err)
			return
		}
		defer outputFile.Close()
		resultsWriter = outputFile
	}

	if len(pDatasetFiles) > 0 {
		for _, filePath := range pDatasetFiles {
			logger.Infof("searching in %s", filePath)

			file, err := os.Open(filePath)
			if err != nil {
				logger.Err(err)
				return
			}

			gzipReader, err := pgzip.NewReaderN(file, pgzipBlockSize, pgzipBlocks)
			if err != nil {
				logger.Err(err)
				return
			}

			resChan, errChan, err := searcher.Search(ctx, search.Options{
				DatasetReader: gzipReader,
				Domains:       pSearchDomains,
				Exclusions:    pExcludedDomains,
				Types:         pSearchTypes,
			})
			if err != nil {
				logger.Err(err)
				return
			}

			go func() {
				for err := range errChan {
					logger.Err(err)
				}
			}()

			for res := range resChan {
				logger.Resultf("%s", res.Name)
				fmt.Fprintln(resultsWriter, res.Name)
			}

			gzipReader.Close()
			file.Close()
		}
	} else {
		datasets, err := dataset.FetchDatasets(ctx)
		if err != nil {
			logger.Err(err)
			return
		}

		selectedDatasets := make([]dataset.Dataset, 0)
		if !pAnyOnly {
			for _, searchType := range pSearchTypes {
				for _, ds := range datasets {
					if ds.HasType(searchType) {
						logger.Infof("selected dataset %s", ds.URL)
						selectedDatasets = append(selectedDatasets, ds)
					}
				}
			}
		}

		if pAny || pAnyOnly {
			for _, ds := range datasets {
				if ds.HasType("any") {
					logger.Infof("selected dataset %s", ds.URL)
					selectedDatasets = append(selectedDatasets, ds)
					break
				}
			}
		}

		if len(selectedDatasets) == 0 {
			logger.Errorf("no matching datasets for types %v found", pSearchTypes)
			return
		}

		for _, selectedDataset := range selectedDatasets {
			logger.Infof("searching in %s", selectedDataset.URL)

			res, err := http.Get(selectedDataset.URL)
			if err != nil {
				logger.Err(err)
				return
			}

			gzipReader, err := pgzip.NewReaderN(res.Body, pgzipBlockSize, pgzipBlocks)
			if err != nil {
				logger.Err(err)
				return
			}

			resChan, errChan, err := searcher.Search(ctx, search.Options{
				DatasetReader: gzipReader,
				Domains:       pSearchDomains,
				Exclusions:    pExcludedDomains,
				Types:         pSearchTypes,
			})
			if err != nil {
				logger.Err(err)
				return
			}

			go func() {
				for err := range errChan {
					logger.Err(err)
				}
			}()

			for res := range resChan {
				logger.Resultf("%s", res.Name)
				fmt.Fprintln(resultsWriter, res.Name)
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
