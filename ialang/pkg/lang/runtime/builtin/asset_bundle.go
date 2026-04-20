package builtin

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// assetBundleModule 创建静态资源打包模块
func newAssetBundleModule() Value {
	// bundle 方法 - 打包资源
	bundleFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("asset.bundle expects at least 2 args, got %d", len(args))
		}

		// 获取参数
		outputDir, err := asStringValue("asset.bundle.outputDir", args[0])
		if err != nil {
			return nil, err
		}

		optionsObj, ok := args[1].(Object)
		if !ok {
			return nil, fmt.Errorf("asset.bundle expects options object as second arg")
		}

		// 解析选项
		files := []string{}
		if filesVal, ok := optionsObj["files"]; ok {
			if filesArr, ok := filesVal.(Array); ok {
				for _, f := range filesArr {
					if s, ok := f.(string); ok {
						files = append(files, s)
					}
				}
			}
		}

		minify := true
		if minifyVal, ok := optionsObj["minify"]; ok {
			if b, ok := minifyVal.(bool); ok {
				minify = b
			}
		}

		hash := true
		if hashVal, ok := optionsObj["hash"]; ok {
			if b, ok := hashVal.(bool); ok {
				hash = b
			}
		}

		concat := false
		if concatVal, ok := optionsObj["concat"]; ok {
			if b, ok := concatVal.(bool); ok {
				concat = b
			}
		}

		// 创建输出目录
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create output directory: %w", err)
		}

		// 处理文件
		manifest := make(Object)
		processedFiles := make(Array, 0)

		for _, file := range files {
			// 读取文件
			content, err := os.ReadFile(file)
			if err != nil {
				return nil, fmt.Errorf("failed to read file %s: %w", file, err)
			}

			// 获取文件类型
			ext := strings.ToLower(filepath.Ext(file))
			contentStr := string(content)

			// 压缩
			if minify {
				switch ext {
				case ".css":
					contentStr = minifyCSS(contentStr)
				case ".js":
					contentStr = minifyJS(contentStr)
				case ".html", ".htm":
					contentStr = minifyHTML(contentStr)
				}
			}

			// 生成哈希
			fileHash := ""
			if hash {
				hashVal := sha256.Sum256([]byte(contentStr))
				fileHash = hex.EncodeToString(hashVal[:])[:8]
			}

			// 生成输出文件名
			baseName := strings.TrimSuffix(filepath.Base(file), ext)
			var outputFileName string
			if hash && fileHash != "" {
				outputFileName = fmt.Sprintf("%s.%s%s", baseName, fileHash, ext)
			} else {
				outputFileName = filepath.Base(file)
			}

			outputPath := filepath.Join(outputDir, outputFileName)

			// 写入文件
			if err := os.WriteFile(outputPath, []byte(contentStr), 0644); err != nil {
				return nil, fmt.Errorf("failed to write file %s: %w", outputPath, err)
			}

			// 添加到清单
			manifest[file] = Object{
				"output":    outputFileName,
				"hash":      fileHash,
				"size":      float64(len(contentStr)),
				"original":  float64(len(content)),
			}

			processedFiles = append(processedFiles, Object{
				"input":     file,
				"output":    outputFileName,
				"hash":      fileHash,
				"size":      float64(len(contentStr)),
				"minified":  minify,
			})
		}

		// 如果需要合并
		if concat && len(files) > 0 {
			ext := strings.ToLower(filepath.Ext(files[0]))
			var mergedContent strings.Builder
			var mergedFiles []string

			for _, file := range files {
				if strings.ToLower(filepath.Ext(file)) == ext {
					content, err := os.ReadFile(file)
					if err != nil {
						continue
					}

					contentStr := string(content)
					if minify {
						switch ext {
						case ".css":
							contentStr = minifyCSS(contentStr)
						case ".js":
							contentStr = minifyJS(contentStr)
						}
					}

					mergedContent.WriteString(contentStr)
					mergedContent.WriteString("\n")
					mergedFiles = append(mergedFiles, file)
				}
			}

			// 生成合并文件
			hashVal := sha256.Sum256([]byte(mergedContent.String()))
			fileHash := hex.EncodeToString(hashVal[:])[:8]

			var mergedFileName string
			if hash {
				mergedFileName = fmt.Sprintf("bundle.%s%s", fileHash, ext)
			} else {
				mergedFileName = fmt.Sprintf("bundle%s", ext)
			}

			mergedPath := filepath.Join(outputDir, mergedFileName)
			if err := os.WriteFile(mergedPath, []byte(mergedContent.String()), 0644); err != nil {
				return nil, fmt.Errorf("failed to write merged file: %w", err)
			}

			manifest["__bundle__"] = Object{
				"output": mergedFileName,
				"hash":   fileHash,
				"size":   float64(mergedContent.Len()),
				"files":  mergedFiles,
			}
		}

		return Object{
			"manifest": manifest,
			"files":    processedFiles,
			"output":   outputDir,
			"count":    float64(len(files)),
		}, nil
	})

	// minify 方法 - 压缩代码
	minifyFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("asset.minify expects 2 args, got %d", len(args))
		}

		content, err := asStringValue("asset.minify.content", args[0])
		if err != nil {
			return nil, err
		}

		fileType, err := asStringValue("asset.minify.type", args[1])
		if err != nil {
			return nil, err
		}

		var result string
		switch strings.ToLower(fileType) {
		case "css":
			result = minifyCSS(content)
		case "js", "javascript":
			result = minifyJS(content)
		case "html", "htm":
			result = minifyHTML(content)
		default:
			return nil, fmt.Errorf("unsupported file type: %s", fileType)
		}

		return Object{
			"original":  float64(len(content)),
			"minified":  float64(len(result)),
			"content":   result,
			"ratio":     float64(len(result)) / float64(len(content)) * 100,
		}, nil
	})

	// hash 方法 - 生成内容哈希
	hashFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("asset.hash expects 1 arg, got %d", len(args))
		}

		content, err := asStringValue("asset.hash.content", args[0])
		if err != nil {
			return nil, err
		}

		length := 8
		if len(args) > 1 {
			if l, ok := args[1].(float64); ok {
				length = int(l)
			}
		}

		hashVal := sha256.Sum256([]byte(content))
		fullHash := hex.EncodeToString(hashVal[:])

		if length > len(fullHash) {
			length = len(fullHash)
		}

		return Object{
			"hash":   fullHash[:length],
			"full":   fullHash,
			"length": float64(length),
		}, nil
	})

	// version 方法 - 为文件添加版本
	versionFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("asset.version expects 1 arg, got %d", len(args))
		}

		filePath, err := asStringValue("asset.version.filePath", args[0])
		if err != nil {
			return nil, err
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		hashVal := sha256.Sum256(content)
		fileHash := hex.EncodeToString(hashVal[:])[:8]

		ext := filepath.Ext(filePath)
		baseName := strings.TrimSuffix(filepath.Base(filePath), ext)

		return Object{
			"path":     filePath,
			"hash":     fileHash,
			"versioned": fmt.Sprintf("%s.%s%s", baseName, fileHash, ext),
			"query":    fmt.Sprintf("%s?v=%s", filepath.Base(filePath), fileHash),
		}, nil
	})

	// clean 方法 - 清理目录
	cleanFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("asset.clean expects 1 arg, got %d", len(args))
		}

		dir, err := asStringValue("asset.clean.dir", args[0])
		if err != nil {
			return nil, err
		}

		pattern := "*"
		if len(args) > 1 {
			if p, ok := args[1].(string); ok {
				pattern = p
			}
		}

		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			return nil, fmt.Errorf("failed to glob pattern: %w", err)
		}

		removed := make(Array, 0)
		for _, match := range matches {
			if err := os.Remove(match); err == nil {
				removed = append(removed, match)
			}
		}

		return Object{
			"removed": removed,
			"count":   float64(len(removed)),
		}, nil
	})

	// analyze 方法 - 分析资源
	analyzeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("asset.analyze expects 1 arg, got %d", len(args))
		}

		dir, err := asStringValue("asset.analyze.dir", args[0])
		if err != nil {
			return nil, err
		}

		stats := Object{
			"totalSize":    float64(0),
			"fileCount":    float64(0),
			"byExtension":  make(Object),
			"files":        make(Array, 0),
		}

		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}

			ext := strings.ToLower(filepath.Ext(path))
			size := info.Size()

			// 更新统计
			totalSize := stats["totalSize"].(float64)
			stats["totalSize"] = totalSize + float64(size)

			fileCount := stats["fileCount"].(float64)
			stats["fileCount"] = fileCount + 1

			// 按扩展名统计
			byExt := stats["byExtension"].(Object)
			if _, ok := byExt[ext]; !ok {
				byExt[ext] = Object{
					"count": float64(0),
					"size":  float64(0),
				}
			}

			extStats := byExt[ext].(Object)
			extStats["count"] = extStats["count"].(float64) + 1
			extStats["size"] = extStats["size"].(float64) + float64(size)

			// 添加文件信息
			files := stats["files"].(Array)
			files = append(files, Object{
				"path": path,
				"size": float64(size),
				"ext":  ext,
			})
			stats["files"] = files

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("failed to analyze directory: %w", err)
		}

		return stats, nil
	})

	return Object{
		"bundle":  bundleFn,
		"minify":  minifyFn,
		"hash":    hashFn,
		"version": versionFn,
		"clean":   cleanFn,
		"analyze": analyzeFn,
	}
}

// minifyCSS 压缩 CSS
func minifyCSS(css string) string {
	// 移除注释
	re := regexp.MustCompile(`/\*.*?\*/`)
	css = re.ReplaceAllString(css, "")

	// 移除多余空白
	re = regexp.MustCompile(`\s+`)
	css = re.ReplaceAllString(css, " ")

	// 移除空格
	css = strings.Replace(css, " {", "{", -1)
	css = strings.Replace(css, "{ ", "{", -1)
	css = strings.Replace(css, " }", "}", -1)
	css = strings.Replace(css, "; ", ";", -1)
	css = strings.Replace(css, ": ", ":", -1)
	css = strings.Replace(css, ", ", ",", -1)

	return strings.TrimSpace(css)
}

// minifyJS 压缩 JavaScript
func minifyJS(js string) string {
	// 移除单行注释（但保留 URL 中的 //）
	re := regexp.MustCompile(`(?m)^\s*//.*$`)
	js = re.ReplaceAllString(js, "")

	// 移除多行注释
	re = regexp.MustCompile(`/\*.*?\*/`)
	js = re.ReplaceAllString(js, "")

	// 移除行尾空白
	re = regexp.MustCompile(`\s+$`)
	js = re.ReplaceAllString(js, "")

	// 移除多余空白（但保留字符串中的）
	re = regexp.MustCompile(`([^\s])\s+([^\s])`)
	js = re.ReplaceAllString(js, "$1 $2")

	return strings.TrimSpace(js)
}

// minifyHTML 压缩 HTML
func minifyHTML(html string) string {
	// 移除注释（保留条件注释）
	re := regexp.MustCompile(`<!--[^[].*?-->`)
	html = re.ReplaceAllString(html, "")

	// 移除标签间的空白
	re = regexp.MustCompile(`>\s+<`)
	html = re.ReplaceAllString(html, "><")

	// 移除属性值周围的空白
	re = regexp.MustCompile(`\s+`)
	html = re.ReplaceAllString(html, " ")

	// 移除空格
	html = strings.Replace(html, " >", ">", -1)
	html = strings.Replace(html, "> ", ">", -1)
	html = strings.Replace(html, "= \"", "=\"", -1)
	html = strings.Replace(html, "\" ", "\"", -1)

	return strings.TrimSpace(html)
}
