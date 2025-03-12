<div align="center">
  <img src="https://github.com/ovftank.png" alt="Logo" width="200" height="200" style="border-radius: 50%;">

# wasm test

  <p align="center">
    <a href="https://ovftank.github.io/wasm-test">Xem Demo</a>
  </p>
</div>

1. **Clone repository**

~~~bash
git clone https://github.com/ovftank/wasm-test.git
cd wasm-test
~~~

2. **Build**

~~~bash
# Trên Linux
GOOS=js GOARCH=wasm go build -o main.wasm

# Trên Windows
set GOOS=js
set GOARCH=wasm
go build -o main.wasm
~~~
