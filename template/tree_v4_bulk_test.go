package template

import (
	"bufio"
	"fmt"
	"github.com/kentik/patricia"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"strings"
	"testing"
)

func TestBulkLoad(t *testing.T) {
	// Test bulk loading
	filePath := "./test_tags.tsv"
	recordsToLoad := -1 // -1 == all

	tree := NewTreeV4(20000000)
	ipToTags := make(map[string]string)
	ips := make([]string, 0)

	// insert the tags
	load := func() {
		file, err := os.Open(filePath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		fmt.Printf("Loading up to %d tags:\n", recordsToLoad)
		scanner := bufio.NewScanner(file)
		recordsLoaded := 0
		for scanner.Scan() {
			if recordsToLoad > 0 {
				if recordsLoaded == recordsToLoad {
					fmt.Printf("Loaded %d: done\n", recordsToLoad)
					break
				}
			}
			recordsLoaded++

			line := scanner.Text()
			parts := strings.Split(line, "\t")
			if len(parts) != 2 {
				panic(fmt.Sprintf("Line should have 2 parts: %s\n", line))
			}

			ipToTags[parts[0]] = parts[1]
			ips = append(ips, parts[0])

			v4, v6, err := patricia.ParseIPFromString(parts[0])
			if err != nil {
				panic(fmt.Sprintf("insert: Could not parse IP '%s': %s", parts[0], err))
			}
			if v4 != nil {
				tree.Add(v4, parts[1])
				continue
			}
			if v6 == nil {
				panic(fmt.Sprintf("insert: Didn't get v4 or v6 address from line: '%s'", line))
			}
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("done - loaded %d tags\n", recordsLoaded)
		fmt.Printf("tree says loaded %d tags into %d nodes\n", tree.countTags(1), tree.countNodes(1))
		assert.Equal(t, uint(recordsLoaded), tree.countTags(1))
	}

	load()
	//tree.print()

	// query all tags from each address, query specific tag from each address, delete the tag
	for _, address := range ips {
		tag := ipToTags[address]
		v4Original, v6, err := patricia.ParseIPFromString(address)
		if err != nil {
			panic(fmt.Sprintf("search: Could not parse IP '%s': %s", address, err))
		}
		if v4Original != nil {
			v4 := *v4Original
			foundTags, err := tree.FindTags(&v4)
			assert.NoError(t, err)
			if assert.True(t, len(foundTags) > 0) {
				assert.True(t, tag == foundTags[len(foundTags)-1])
			}

			v4 = *v4Original
			found, foundTag, err := tree.FindDeepestTag(&v4)
			assert.NoError(t, err)
			assert.True(t, found, "Couldn't find deepest tag")
			assert.True(t, tag == foundTag)

			// delete the tags now
			//fmt.Printf("Deleting %s: %s\n", address, tag)
			v4 = *v4Original
			deleteCount, err := tree.Delete(&v4, func(a GeneratedType, b GeneratedType) bool { return a == b }, tag)
			assert.NoError(t, err)
			assert.Equal(t, 1, deleteCount, "Tried deleting tag")
			//tree.print()
		} else if v6 == nil {
			panic(fmt.Sprintf("search: Didn't get v4 or v6 address from address: '%s'", address))
		}
	}

	// should be nothing left
	assert.Equal(t, uint(0), tree.countTags(1))
	assert.Equal(t, 1, tree.countNodes(1))

	fmt.Printf("Finished looping, finding, deleting - Tree now has %d tags in %d nodes\n", tree.countTags(1), tree.countNodes(1))

	//load()
	//print()
}

func arraysEqual(tag string, expected []string, found []GeneratedType) bool {
	if len(expected) != len(found) {
		return false
	}

	for _, tagA := range expected {
		didFind := false
		for _, tagB := range found {
			if tagB == tagA {
				didFind = true
				break
			}
		}
		if !didFind {
			return false
		}
	}
	return true
}
