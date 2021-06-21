# DataArt Go Client

DataArt platform client for Golang.

## Getting Started

```bash
  $ go get -u github.com/dataart-ai/dataart-go
```

## Example

```go
package main

import (
	"net/http"
	"time"

	"github.com/dataart-ai/dataart-go/pkg/dataart"
)

func main() {
	cfg := dataart.ClientConfig{
		APIKey:        "your-api-key",
		FlushCap:      20,
		FlushInterval: time.Duration(30 * time.Second),
		HTTPClient:    http.DefaultClient,
	}

	c, err := dataart.DefaultClient(cfg)
	if err != nil {
		// error handling...
		return
	}
	defer c.Close()

	err = c.EmitAction("some-event-key", "some-user-key", false, time.Now(), map[string]interface{}{
		"metadata_key_1": "metadata_value_1",
		"metadata_key_2": 2,
	})
	if err != nil {
		// error handling
		return
	}

	err = c.Identify("some-user-key", map[string]interface{}{
		"metadata_key_3": "metadata_value_3",
		"metadata_key_4": 4,
	})
	if err != nil {
		// error handling
		return
	}
}

```

## License

MIT License

Copyright (c) 2021 DataArt

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
