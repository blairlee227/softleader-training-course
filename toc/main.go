package main

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"text/template"
)

const (
	coursesTemplate = `> Auto-generated by *toc*
{{- $years := index . "years" }}
{{- $courses := index . "courses" }}
{{- range $_, $year := $years }}

#### {{ $year -}}

{{- range $_, $season := index $courses $year }}
{{- range $_, $course := $season.Courses }}
- {{ $season.Season }} - [{{ $course.Course }}](/{{ $course.Year }}/{{ $course.Season }}/{{ $course.Course }}){{ if ne $course.Description "" }} - {{ $course.Description }}{{- end }}
{{- end }}
{{- end }}
{{- end }}`
)

type config struct {
	workdir  string
	scandir  string
	template string
	output   string
}

func main() {
	c := &config{}

	cmd := &cobra.Command{
		Use:   "toc",
		Short: "Table Of Content generator",
		Long: `Table Of Content Generator

	$ toc PATH

PATH 用來指定從哪層目錄 (相對於工作目錄) 開始爬文, 如: '.' 代表當前目錄

	$ toc .

傳入 '--workdir' 指定工作目錄, 可為絕對路徑或相對於當前目錄的路徑, 預設執行指令的當前目錄

	$ toc . --workdir ../../
	$ toc . --workdir /tmp

傳入 '--template' 指定 template 位置 (相對於工作目錄)

	$ toc . --template templates/my.tpl

傳入 '--output' 指令輸出的檔案

	$ toc . --template templates/my.tpl --output my.file
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c.scandir = args[0]
			return generateTOC(c)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&c.workdir, "workdir", "", "working directory (default current directory)")
	flags.StringVarP(&c.output, "output", "o", "./README.md", "output file name relative to workdir")
	flags.StringVarP(&c.template, "template", "t", "", "template file to use relative to workdir (required)")

	cmd.MarkFlagRequired("template")

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func generateTOC(c *config) (err error) {
	if c.workdir, err = filepath.Abs(c.workdir); err != nil {
		return err
	}
	fmt.Println(c.workdir)

	return

	// 爬子目錄收集所有課程目錄
	var courses []Course
	if err = walkDir(filepath.Join(c.workdir, c.scandir), 0, 2, func(path string) {
		season := filepath.Dir(path)
		year := filepath.Dir(season)
		if y, err := strconv.Atoi(filepath.Base(year)); err == nil { // 預期"年"一定要可以轉數字, 避免載入太多不相干的目錄
			c := Course{
				Year:   y,
				Season: filepath.Base(season),
				Course: filepath.Base(path),
			}
			if md, err := ioutil.ReadFile(filepath.Join(path, "README.md")); err == nil {
				re := regexp.MustCompile("#{1,6}\\s+(.+)")
				match := re.FindStringSubmatch(string(md))
				if len(match) > 0 {
					if desc := match[1]; desc != c.Course {
						c.Description = match[1]
					}
				}
			}
			courses = append(courses, c)
		}
	}); err != nil {
		return
	}

	// 將課程轉換成 template 容易的結構
	var years []int
	groupByYear := make(map[int]Seasons)
	for _, c := range courses {
		if seasons, found := groupByYear[c.Year]; found {
			seasons.add(c)
			groupByYear[c.Year] = seasons
		} else {
			seasons = Seasons{}
			seasons.add(c)
			groupByYear[c.Year] = seasons
			years = append(years, c.Year)
		}
	}

	// 排序年, 越舊的往後, 越新的往前
	sort.Ints(years)
	for i := len(years)/2 - 1; i >= 0; i-- {
		opp := len(years) - 1 - i
		years[i], years[opp] = years[opp], years[i]
	}

	// 套用 courses template
	data := make(map[string]interface{})
	data["years"] = years
	data["courses"] = groupByYear
	renderedCourses, err := renderTemplate(coursesTemplate, data)
	if err != nil {
		return err
	}

	// 套用 readme template
	readmeTemplate, err := ioutil.ReadFile(filepath.Join(c.workdir, c.template))
	if err != nil {
		return err
	}
	renderedReadme, err := renderTemplate(string(readmeTemplate), renderedCourses)
	if err != nil {
		return err
	}

	// 輸出 readme
	readme := filepath.Join(c.workdir, c.output)
	ioutil.WriteFile(readme, []byte(renderedReadme), os.ModePerm)

	fmt.Printf("Successfully genereated %s\n", readme)
	return
}

func renderTemplate(text string, data interface{}) (string, error) {
	var buf bytes.Buffer
	t := template.Must(template.New("").Parse(text))
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// 在第 depth 層的時候依序針對當下的 dir 執行 walkFn
func walkDir(dirpath string, currentDepth int, depth int, walkFn func(path string)) error {
	if currentDepth > depth {
		return nil
	}
	files, err := ioutil.ReadDir(dirpath)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			p := path.Join(dirpath, file.Name())
			if currentDepth == depth {
				walkFn(p)
			}
			walkDir(p, currentDepth+1, depth, walkFn)
			continue
		} else {
			continue
		}
	}
	return nil
}

type Seasons []Season

func (seasons *Seasons) add(c Course) {
	for i := 0; i < len(*seasons); i++ {
		if (*seasons)[i].Season == c.Season {
			(*seasons)[i].Courses = append((*seasons)[i].Courses, c)
			return
		}
	}
	*seasons = append(*seasons, Season{
		Season:  c.Season,
		Courses: []Course{c},
	})
}

type Season struct {
	Season  string
	Courses []Course
}

type Course struct {
	Year        int
	Season      string
	Course      string
	Description string
}
