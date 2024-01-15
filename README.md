## Usage

```bash
 go run ./main.go -batch 2 \
        -file input_text \
        -file single_word \
        -file https://raw.githubusercontent.com/romashorodok/music-service/main/README.md \
        -file compiled_rust_hello_world
```
Output:
```
2024/01/15 19:12:59 Running stdlib approach? false
2024/01/15 19:12:59 Skip file `compiled_rust_hello_world` because it's executable!
2024/01/15 19:12:59 single_word:  [>9] = 1
2024/01/15 19:12:59 single_word:  total words: 1
2024/01/15 19:12:59 single_word:  took 162.162µs
2024/01/15 19:12:59 input_text:  [2] = 13, [3] = 5, [4] = 9, [5] = 12, [6] = 7, [7] = 9, [8] = 3, [9] = 5, [>9] = 6
2024/01/15 19:12:59 input_text:  total words: 69
2024/01/15 19:12:59 input_text:  took 463.846µs
2024/01/15 19:12:59 https://raw.githubusercontent.com/romashorodok/music-service/main/README.md:  [2] = 8, [3] = 11, [4] = 2, [5] = 15, [6] = 11, [7] = 9, [8] = 1, [9] = 5, [>9] = 22
2024/01/15 19:12:59 https://raw.githubusercontent.com/romashorodok/music-service/main/README.md:  total words: 84
2024/01/15 19:12:59 https://raw.githubusercontent.com/romashorodok/music-service/main/README.md:  took 138.972731ms
2024/01/15 19:12:59 Process: took 139.262431ms
```
