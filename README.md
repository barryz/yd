**yd** 是一个用于翻译单词的命令行工具。

# 安装
```bash
go get -u github.com/barryz/yd
```

# 二进制下载

暂无

# 使用

### 仅翻译单词

```sh
yd -w $word
```

### 翻译结果导入 [Anki](https://apps.ankiweb.net/) 中

前提是本机已安装 Anki ， 且配置好 AnkiConnect Add-ons。

```bash
export ANKI_DECK_NAME=$your_anki_deck_name # 你的 anki deck 名称
yd -w $word -anki
```

为了简化，可以使用 alias :

```bash
echo 'alias tra="ANKI_DECK_NAME=$your_anki_deck_name yd -anki -w"' >> ~/.zshrc
# or
echo 'alias tra="ANKI_DECK_NAME=$your_anki_deck_name yd -anki -w"' >> ~/.bashrc

# import the result into Anki when you query the word
tra $word
```

### Anki 相关

- [Anki](https://apps.ankiweb.net/)
- [AnkiConnect Add-ons](https://ankiweb.net/shared/info/2055492159)

### TODO

- [ ] 支持文本翻译
- [ ] 支持汉英翻译
- [ ] 其他
