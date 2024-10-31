(function_declaration
    "async"? @context
    "function" @context
    name: (_) @name
    parameters: (formal_parameters
      "(" @context
      ")" @context)) @item


(program
    (export_statement
        (lexical_declaration
            ["let" "const"] @context
            (variable_declarator
                name: (_) @name) @item)))

(program
    (lexical_declaration
        ["let" "const"] @context
        (variable_declarator
            name: (_) @name) @item))

(class_declaration
    "class" @context
    name: (_) @name) @item

(method_definition
    [
        "get"
        "set"
        "async"
        "*"
        "static"
    ]* @context
    name: (_) @name
    parameters: (formal_parameters
      "(" @context
      ")" @context)) @item


; Add support for (node:test, bun:test and Jest) runnable
(call_expression
    function: (_) @context
    (#any-of? @context "it" "test" "describe")
    arguments: (
        arguments . (string
            (string_fragment) @name
        )
    )
) @item

(comment) @annotation
