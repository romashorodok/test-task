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
2024/01/15 10:11:16 Running stdlib approach? false
2024/01/15 10:11:16 Skip file `compiled_rust_hello_world` because it's executable!
2024/01/15 10:11:16 single_word:  [>9] = 1
2024/01/15 10:11:16 single_word:  word: 1
2024/01/15 10:11:16 single_word:  took 121.725µs
2024/01/15 10:11:16 input_text:  [2] = 13, [3] = 5, [4] = 9, [5] = 12, [6] = 7, [7] = 9, [8] = 3, [9] = 5, [>9] = 6
2024/01/15 10:11:16 input_text:  word: 69
2024/01/15 10:11:16 input_text:  took 402.31µs
2024/01/15 10:11:16 https://raw.githubusercontent.com/romashorodok/music-service/main/README.md:  [2] = 8, [3] = 11, [4] = 2, [5] = 15, [6] = 11, [7] = 9, [8] = 1, [9] = 5, [>9] = 22
2024/01/15 10:11:16 https://raw.githubusercontent.com/romashorodok/music-service/main/README.md:  word: 84
2024/01/15 10:11:16 https://raw.githubusercontent.com/romashorodok/music-service/main/README.md:  took 135.435742ms
2024/01/15 10:11:16 Process: took 135.606534ms
```
