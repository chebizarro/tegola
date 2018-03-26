package atlas_test

import (
	"testing"
	"reflect"
	"context"

	"github.com/go-spatial/tegola"
	"github.com/go-spatial/tegola/atlas"
	"github.com/go-spatial/tegola/geom"
	"github.com/go-spatial/tegola/provider/test"
	"github.com/go-spatial/tegola/cache/memory"
)

var testLayer1 = atlas.Layer{
	Name:              "test-layer",
	ProviderLayerName: "test-layer-1",
	MinZoom:           4,
	MaxZoom:           9,
	Provider:          &test.TileProvider{},
	GeomType:          geom.Point{},
	DefaultTags: map[string]interface{}{
		"foo": "bar",
	},
}

var testLayer2 = atlas.Layer{
	Name:              "test-layer-2-name",
	ProviderLayerName: "test-layer-2-provider-layer-name",
	MinZoom:           10,
	MaxZoom:           20,
	Provider:          &test.TileProvider{},
	GeomType:          geom.LineString{},
	DefaultTags: map[string]interface{}{
		"foo": "bar",
	},
}

var testLayer3 = atlas.Layer{
	Name:              "test-layer",
	ProviderLayerName: "test-layer-3",
	MinZoom:           10,
	MaxZoom:           20,
	Provider:          &test.TileProvider{},
	GeomType:          geom.Point{},
	DefaultTags:       map[string]interface{}{},
}

var testMap = atlas.Map{
	Name:        "test-map",
	Attribution: "test attribution",
	Center:      [3]float64{1.0, 2.0, 3.0},
	Layers: []atlas.Layer{
		testLayer1,
		testLayer2,
		testLayer3,
	},
}


func TestAtlasAddMap(t *testing.T) {

	testAtlas := new(atlas.Atlas)
	testAtlas.AddMap(testMap)
	
	atlasMaps := reflect.Indirect(reflect.ValueOf(testAtlas))
	mapCount := atlasMaps.FieldByName("maps").Len()
	
	if mapCount != 1 {
		t.Errorf("Number of maps in the Atlas was incorrect, got: %d, want: %d.", mapCount, 1)
	}
}

func TestAtlasAllMaps(t *testing.T) {
	
	testAtlas := new(atlas.Atlas)
	testAtlas.AddMap(testMap)

	allMaps := testAtlas.AllMaps()
	mapCount := len(allMaps)
	
	if mapCount != 1 {
		t.Errorf("Number of maps in the Atlas was incorrect, got: %d, want: %d.", mapCount, 1)
	}	
}

func TestAtlasMap(t *testing.T) {
	
	testAtlas := new(atlas.Atlas)
	testAtlas.AddMap(testMap)

	_, ok := testAtlas.Map("test-map")
	
	if ok != nil {
		t.Errorf(ok.Error())
	}
}

func TestAtlasMapNotFound(t *testing.T) {
	
	testAtlas := new(atlas.Atlas)
	testAtlas.AddMap(testMap)
	mapName := "does-not-exist"
	
	_, ok := testAtlas.Map(mapName)
	
	if ok == nil {
		t.Errorf("Atlas should not contain a map named: %s",mapName)
	}
}

func TestAtlasSeedMapTile(t *testing.T) {
	
	testAtlas := new(atlas.Atlas)
	testAtlas.AddMap(testMap)
	testAtlas.SetCache(memory.New())

	if ok := testAtlas.SeedMapTile(context.Background(), testMap, 0,0,0); ok != nil {
		t.Errorf(ok.Error())	
	}
	
}

func TestAtlasPurgeMapTile(t *testing.T) {
	
	testAtlas := new(atlas.Atlas)
	testAtlas.AddMap(testMap)
	testAtlas.SetCache(memory.New())

	if ok := testAtlas.PurgeMapTile(testMap, new(tegola.Tile)); ok != nil {
		t.Errorf(ok.Error())	
	}
	
}
