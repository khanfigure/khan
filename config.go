package duck

/*import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func Config(target interface{}) {
	if err := loadConfig(target); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func loadConfig(target interface{}) error {
	fpath := "co.yaml"

	fmt.Println("reading", fpath)

	buf, err := ioutil.ReadFile(fpath)
	if err != nil {
		return errors.Wrapf(err, "Read %#v", fpath)
	}
	if err := yaml.Unmarshal(buf, target); err != nil {
		return errors.Wrapf(err, "Parse %#v", fpath)
	}

	return nil
}*/
