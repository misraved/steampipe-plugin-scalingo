package scalingo

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/Scalingo/go-scalingo/v4"
	"github.com/turbot/steampipe-plugin-sdk/v2/connection"
	"github.com/turbot/steampipe-plugin-sdk/v2/plugin"
)

const matrixKeyRegion = "region"
const defaultScalingoRegion = "osc-fr1"

var pluginQueryData *plugin.QueryData

func init() {
	pluginQueryData = &plugin.QueryData{
		ConnectionManager: connection.NewManager(),
	}
}

func connect(ctx context.Context, d *plugin.QueryData) (*scalingo.Client, error) {
	region := d.KeyColumnQualString(matrixKeyRegion)
	if region == "" {
		region = defaultScalingoRegion
	}

	// get scalingo client from cache
	cacheKey := fmt.Sprintf("scalingo-%s", region)
	if cachedData, ok := d.ConnectionManager.Cache.Get(cacheKey); ok {
		return cachedData.(*scalingo.Client), nil
	}

	token := os.Getenv("SCALINGO_TOKEN")

	scalingoConfig := GetConfig(d.Connection)
	if &scalingoConfig != nil {
		if scalingoConfig.Token != nil {
			token = *scalingoConfig.Token
		}
	}

	if token == "" {
		return nil, errors.New("'token' must be set in the connection configuration. Edit your connection configuration file or set the SCALINGO_TOKEN environment variable and then restart Steampipe")
	}

	config := scalingo.ClientConfig{
		APIToken: token,
		Region:   region,
	}
	client, err := scalingo.New(config)

	if err != nil {
		return nil, err
	}

	// Save to cache
	d.ConnectionManager.Cache.Set(cacheKey, client)

	return client, nil
}

func BuildRegionList(ctx context.Context, connection *plugin.Connection) []map[string]interface{} {
	pluginQueryData.Connection = connection

	// cache matrix
	cacheKey := "RegionListMatrix"
	if cachedData, ok := pluginQueryData.ConnectionManager.Cache.Get(cacheKey); ok {
		return cachedData.([]map[string]interface{})
	}

	var regions []string

	// retrieve regions from connection config
	scalingoConfig := GetConfig(connection)

	if &scalingoConfig != nil {
		// handle compatibility with the old region configuration
		if scalingoConfig.Region != nil {
			regions = append(regions, *scalingoConfig.Region)
		}

		// Get only the regions as required by config file
		if len(*scalingoConfig.Regions) > 0 {
			regions = *scalingoConfig.Regions
		}
	}

	if len(regions) > 0 {
		matrix := make([]map[string]interface{}, len(regions))
		for i, region := range regions {
			matrix[i] = map[string]interface{}{matrixKeyRegion: region}
		}

		// set cache
		pluginQueryData.ConnectionManager.Cache.Set(cacheKey, matrix)
		return matrix
	}

	matrix := []map[string]interface{}{
		{matrixKeyRegion: defaultScalingoRegion},
	}

	// set cache
	pluginQueryData.ConnectionManager.Cache.Set(cacheKey, matrix)
	return matrix
}
