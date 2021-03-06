// Copyright © 2016 Frederick F. Kautz IV fkautz@alumni.cmu.edu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/fkautz/tigerbat/cache/diskcache"
	"github.com/fkautz/tigerbat/cache/httpserver"
	"github.com/fkautz/tigerbat/cache/hydrator"
	"github.com/fkautz/tigerbat/cache/memorycache"
	"github.com/gorilla/mux"
	"github.com/pivotal-golang/bytefmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

// flags
var (
	address          string
	cleanedDiskUsage string
	diskCacheDir     string
	diskCacheEnabled bool
	maxDiskUsage     string
	maxMemoryUsage   string
	mirrorUrl        string
	peeringAddress   string
	etcd             []string
)

func InitializeConfig(cmd *cobra.Command) {
	viper.SetDefault("address", ":8080")
	viper.SetDefault("cleaned-disk-usage", "800M")
	viper.SetDefault("disk-cache-dir", "./data")
	viper.SetDefault("disk-cache-enabled", true)
	viper.SetDefault("max-disk-usage", "1G")
	viper.SetDefault("max-memory-usage", "100M")
	viper.SetDefault("mirror-url", "http://localhost:9000")
	viper.SetDefault("peering-address", "")
	viper.SetDefault("etcd", "")

	if flagChanged(cmd.PersistentFlags(), "address") {
		viper.Set("address", address)
	}
	if flagChanged(cmd.PersistentFlags(), "cleaned-disk-usage") {
		viper.Set("cleaned-disk-usage", cleanedDiskUsage)
	}
	if flagChanged(cmd.PersistentFlags(), "disk-cache-dir") {
		viper.Set("disk-cache-dir", diskCacheDir)
	}
	if flagChanged(cmd.PersistentFlags(), "disk-cache-enabled") {
		viper.Set("disk-cache-enabled", diskCacheEnabled)
	}
	if flagChanged(cmd.PersistentFlags(), "max-disk-usage") {
		viper.Set("max-disk-usage", maxDiskUsage)
	}
	if flagChanged(cmd.PersistentFlags(), "max-meory-usage") {
		viper.Set("max-memory-usage", maxMemoryUsage)
	}
	if flagChanged(cmd.PersistentFlags(), "mirror-url") {
		viper.Set("mirror-url", mirrorUrl)
	}
	if flagChanged(cmd.PersistentFlags(), "peering-address") {
		viper.Set("peering-address", peeringAddress)
	}
	if flagChanged(cmd.PersistentFlags(), "etcd") {
		viper.Set("etcd", etcd)
	}
}

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFlags(log.Flags() | log.Lshortfile)

		InitializeConfig(cmd)
		blockSize := int64(2 * 1024 * 1024)

		var persistentCache diskcache.Cache
		if viper.GetBool("disk-cache-enabled") {
			maxSize, err := bytefmt.ToBytes(viper.GetString("max-disk-usage"))
			if err != nil {
				log.Fatalln("Unable to parse max-disk-usage", err)
			}
			cleanedSize, err := bytefmt.ToBytes(viper.GetString("cleaned-disk-usage"))
			if err != nil {
				log.Fatalln("Unable to parse cleaned-disk-usage", err)
			}
			persistentCache, err = diskcache.New(viper.GetString("disk-cache-dir"), int64(maxSize), int64(cleanedSize))
			if err != nil {
				log.Fatalln("Unable to initialize disk cache", err)
			}
		}

		maxMemory, err := bytefmt.ToBytes(viper.GetString("max-memory-usage"))
		if err != nil {
			log.Fatalln("Unable to parse max-memory-usage", err)
		}

		cacheConfig := gcache.Config{
			MaxMemoryUsage: int64(maxMemory),
			BlockSize:      blockSize,
			DiskCache:      persistentCache,
			Hydrator:       hydrator.NewHydrator(viper.GetString("mirror-url")),
			PeeringAddress: viper.GetString("peering-address"),
			Etcd:           viper.GetStringSlice("etcd"),
		}

		cache := gcache.NewCache(cacheConfig)

		cacheHandler := httpserver.NewHttpHandler(cache, blockSize)

		router := mux.NewRouter()

		groupCacheProxyHandler := http.Handler(cacheHandler)
		router.Handle("/{request:.*}", groupCacheProxyHandler)

		// serve
		//err = http.ListenAndServeTLS(address, "cert.pem", "key.pem", router)
		//handler := lox.NewHandler(lox.NewMemoryCache(), router)
		var handler http.Handler
		handler = router
		//handler = handlers.LoggingHandler(os.Stderr, handler)
		err = http.ListenAndServe(address, handler)
		if err != nil {
			log.Fatalln(err)
		}
	},
}

type appConfig struct {
	urlRoot  string
	listenOn string
}

func init() {
	RootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	serverCmd.PersistentFlags().StringVar(&address, "address", "localhost:8080", "Address to listen on")
	serverCmd.PersistentFlags().StringVar(&maxMemoryUsage, "max-memory-usage", "100M", "Address to listen on")
	serverCmd.PersistentFlags().StringVar(&maxDiskUsage, "max-disk-usage", "1G", "Address to listen on")
	serverCmd.PersistentFlags().StringVar(&cleanedDiskUsage, "cleaned-disk-usage", "800M", "Address to listen on")
	serverCmd.PersistentFlags().BoolVar(&diskCacheEnabled, "disk-cache-enabled", true, "Address to listen on")
	serverCmd.PersistentFlags().StringVar(&diskCacheDir, "disk-cache-dir", "./data", "Address to listen on")
	serverCmd.PersistentFlags().StringVar(&mirrorUrl, "mirror-url", "http://localhost:9000", "URL root to mirror")
	serverCmd.PersistentFlags().StringVar(&peeringAddress, "peering-address", "http://localhost:8000", "URL root to mirror")
	serverCmd.PersistentFlags().StringSliceVar(&etcd, "etcd", []string{}, "URL root to mirror")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
