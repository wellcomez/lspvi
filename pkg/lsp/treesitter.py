from os import read
import tree_sitter

# 初始化JavaScript解析器
go_so="/nvim.plugin/nvim-treesitter/parser/go.so"
lang= tree_sitter.Language(go_so, 'go')
parser=tree_sitter.Parser()

# 解析一段JavaScript代码
codefile = "/home/z/dev/lsp/goui/main.go"
code=open(codefile,"r",encoding="utf-8").read()
highlight_query = "/home/z/dev/lsp/goui/pkg/lsp/queries/go/highlights.scm"
query_scm=open(highlight_query,"r",encoding="utf-8").read()
parser.set_language(lang)
tree = parser.parse(bytes(code, 'utf8'))

query=lang.query(query_scm)
matches = query.captures(tree.root_node)

# 输出匹配结果
for match in matches:
    node, index = match
    func_name = code[node.start_byte:node.end_byte]
    # print(f"Function name: {func_name.decode('utf8')}")
    print(index,node,"\n")
# 打印树结构
print(tree.root_node.sexp())