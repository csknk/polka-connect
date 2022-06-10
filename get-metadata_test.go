package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

func Test_getLatestMetadata(t *testing.T) {
	c, err := NewConnection("http://localhost:9933")
	assert.NoError(t, err)

	meta, err := c.getLatestMetadata()
	assert.NoError(t, err)

	c.getEvents(meta)

}

func TestGetEventsFromMeta(t *testing.T) {
	c, err := NewConnection("http://localhost:9933")
	assert.NoError(t, err)
	meta, err := c.getLatestMetadata()
	assert.NoError(t, err)

	targetMod := types.Text("NominationPools")
	targetVariant := types.Text("StateChanged")
	for _, mod := range meta.AsMetadataV14.Pallets {
		if !mod.HasEvents {
			continue
		}
		// filter mod of interest
		if mod.Name != targetMod {
			continue
		}
		outputTargetedPalletEventData(t, meta, mod, targetVariant)
	}
}

func TestGetAllEvents(t *testing.T) {
	c, err := NewConnection("http://localhost:9933")
	assert.NoError(t, err)
	meta, err := c.getLatestMetadata()
	assert.NoError(t, err)

	for _, mod := range meta.AsMetadataV14.Pallets {
		if !mod.HasEvents {
			continue
		}
		outputAllPalletEventData(t, meta, mod)
	}
}

func outputAllPalletEventData(t *testing.T, meta *types.Metadata, mod types.PalletMetadataV14) {
	t.Helper()
	typ, ok := meta.AsMetadataV14.EfficientLookup[mod.Events.Type.Int64()]
	if !ok {
		return
	}
	typBytes, err := json.MarshalIndent(typ, "", "\t")
	assert.NoError(t, err)
	fmt.Printf("%s\n", typBytes)
	if typ.Def.IsVariant {
		if len(typ.Def.Variant.Variants) == 0 {
			return
		}
		for _, variant := range typ.Def.Variant.Variants {
			outputVariantData(t, mod, variant, meta)
		}
	} else {
		fmt.Printf("Not a variant: %#v\n", typ.Def)
	}
}

func outputTargetedPalletEventData(t *testing.T, meta *types.Metadata, mod types.PalletMetadataV14, targetVariant types.Text) {
	t.Helper()
	typ, ok := meta.AsMetadataV14.EfficientLookup[mod.Events.Type.Int64()]
	if !ok {
		return
	}
	fmt.Printf("---------------------------------------------------------\n")
	if !ok {
		return
	}
	if typ.Def.IsVariant {
		if len(typ.Def.Variant.Variants) == 0 {
			return
		}
		for _, vars := range typ.Def.Variant.Variants {
			if vars.Name != targetVariant {
				continue
			}
			outputVariantData(t, mod, vars, meta)
		}
	}
}

func outputVariantData(t *testing.T, mod types.PalletMetadataV14, vars types.Si1Variant, meta *types.Metadata) {
	fmt.Printf("%s.%s\n", mod.Name, vars.Name)
	varsBytes, err := json.MarshalIndent(vars, "", "\t")
	assert.NoError(t, err)
	fmt.Printf("%s\n", varsBytes)

	outputFields(vars.Fields, meta, "", t)

}

func outputFields(fields []types.Si1Field, meta *types.Metadata, prefix string, t *testing.T) {
	for i, field := range fields {
		name := types.Text(" ")
		if field.HasName {
			name += field.Name
		}
		fmt.Printf("field%s at index %d\n", name, i)
		if subtype, ok := meta.AsMetadataV14.EfficientLookup[field.Type.Int64()]; ok {
			var obj interface{}
			switch definition := subtype.Def; {
			case definition.IsVariant:
				fmt.Println("Variant")
				obj = definition.Variant
			case definition.IsComposite:
				fmt.Println("Composite")
				obj = subtype.Def.Composite
			case definition.IsArray:
				obj = subtype.Def.Array
			case definition.IsBitSequence:
				obj = subtype.Def.BitSequence
			case definition.IsComposite:
				obj = subtype.Def.Composite
			case definition.IsCompact:
				obj = subtype.Def.Compact
			case definition.IsHistoricMetaCompat:
				obj = subtype.Def.HistoricMetaCompat
			case definition.IsPrimitive:
				obj = subtype.Def.Primitive
			case definition.IsTuple:
				obj = subtype.Def.Tuple
			case definition.IsSequence:
				obj = subtype.Def.Sequence
			}
			fieldBytes, err := json.MarshalIndent(obj, prefix, "\t")
			assert.NoError(t, err)
			fmt.Printf("%s\n", fieldBytes)

			// if more fields, send an array of fields back in recursively
			if subtype.Def.IsVariant {
				fmt.Println("looping over Variants...")

				for _, variant := range subtype.Def.Variant.Variants {
					prefix += "\t"
					outputFields(variant.Fields, meta, prefix, t)
				}
			}
			if subtype.Def.IsComposite {
				prefix += "\t"
				outputFields(subtype.Def.Composite.Fields, meta, prefix, t)
			}
		}
	}
}

func outputVariantDataFull(t *testing.T, mod types.PalletMetadataV14, vars types.Si1Variant, meta *types.Metadata) {
	fmt.Printf("%s.%s\n", mod.Name, vars.Name)
	varsBytes, err := json.MarshalIndent(vars, "", "\t")
	assert.NoError(t, err)
	fmt.Printf("%s\n", varsBytes)
	/*
		for _, field := range vars.Fields {
			subtype := meta.AsMetadataV14.Lookup.Types[field.Type.Int64()].Type
			//		var jsonData := map[string]interface{}
			b, err := json.MarshalIndent(subtype.Def, "", "\t")
			assert.NoError(t, err)
			fmt.Printf("%s\n", b)

			fmt.Printf("Field Name: %s Type Name: %s\n", field.Name, field.TypeName)

			if subtype.Def.IsComposite {
				typeName := types.Text("")
				for i, typeField := range subtype.Def.Composite.Fields {
					if typeField.HasTypeName {
						typeName = typeField.TypeName
					}
					fmt.Printf("composite field index %d typeName: %s\n", i, typeName)
					fmt.Printf("typeField: %#v", typeField)
					thisType := meta.AsMetadataV14.EfficientLookup[typeField.Type.UCompact.Int64()]
					fmt.Printf("%#v\n", thisType)
				}
			} else {
				if subtype.Def.IsVariant {
					for _, variant := range subtype.Def.Variant.Variants {
						vBytes, err := json.MarshalIndent(variant, "", "\t")
						assert.NoError(t, err)
						fmt.Println("subtype.Def.Variant.Variants[]:")
						fmt.Printf("%s\n", vBytes)
						for k, f := range variant.Fields {
							fmt.Printf("field %d ++++++\n%#v", k, f)
							x := meta.AsMetadataV14.EfficientLookup[f.Type.Int64()].Def
							xBytes, err := json.MarshalIndent(x, "", "\t")
							assert.NoError(t, err)

							fmt.Printf("%s\n", xBytes)

						}
					}
				}
			}
		}
	*/
}
