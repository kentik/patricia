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

	tree := NewTreeV4()
	var ipToTags map[string]string
	var ips []string

	// insert the tags
	load := func() {
		ipToTags = make(map[string]string)
		ips = make([]string, 0)

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
				tree.Add(*v4, parts[1], nil)
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
		assert.Equal(t, recordsLoaded, int(tree.countTags(1)))
		assert.Equal(t, recordsLoaded, tree.CountTags())
	}

	buf := make([]GeneratedType, 0)
	evaluate := func() {
		fmt.Printf("# of nodes: %d\n", len(tree.nodes))
		// query all tags from each address, query specific tag from each address, delete the tag
		for _, address := range ips {
			tag := ipToTags[address]
			v4, v6, err := patricia.ParseIPFromString(address)
			if err != nil {
				panic(fmt.Sprintf("search: Could not parse IP '%s': %s", address, err))
			}
			if v4 != nil {
				foundTags := tree.FindTags(buf, *v4)
				if assert.True(t, len(foundTags) > 0, "Couldn't find tags for "+address) {
					assert.True(t, tag == foundTags[len(foundTags)-1])
				}

				found, foundTag := tree.FindDeepestTag(*v4)
				assert.True(t, found, "Couldn't find deepest tag")
				assert.True(t, tag == foundTag)

				// delete the tags now
				//fmt.Printf("Deleting %s: %s\n", address, tag)
				deleteCount := tree.Delete(buf, *v4, func(a GeneratedType, b GeneratedType) bool { return a == b }, tag)
				assert.Equal(t, 1, deleteCount, "Tried deleting tag")
				//tree.print()
			} else if v6 == nil {
				panic(fmt.Sprintf("search: Didn't get v4 or v6 address from address: '%s'", address))
			}
		}
		// should be nothing left
		assert.Equal(t, 0, tree.countTags(1))
		assert.Equal(t, 1, tree.countNodes(1))
		fmt.Printf("Finished looping, finding, deleting - Tree now has %d tags in %d logical nodes, %d capacity, %d node objects in use, %d available indexes\n", tree.countTags(1), tree.countNodes(1), cap(tree.nodes), len(tree.nodes), len(tree.availableIndexes))
	}

	//tree.print()
	load()
	evaluate()
	nodesCapacity := cap(tree.nodes)
	nodesLength := len(tree.nodes)
	fmt.Printf("Finished first pass - node capacity: %d\n", cap(tree.nodes))

	// do it again a few times, and make sure the nodes array hasn't grown
	load()
	evaluate()
	load()
	evaluate()
	load()
	evaluate()
	load()
	evaluate()
	assert.Equal(t, nodesLength, len(tree.nodes))
	assert.Equal(t, nodesCapacity, cap(tree.nodes))
	fmt.Printf("Finished looping, finding, deleting - Tree now has %d tags in %d nodes\n", tree.countTags(1), tree.countNodes(1))

	// now try cloning
	load()
	tree = tree.Clone()
	evaluate()
	assert.Equal(t, nodesLength, len(tree.nodes))
	assert.Equal(t, nodesCapacity, cap(tree.nodes))

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
