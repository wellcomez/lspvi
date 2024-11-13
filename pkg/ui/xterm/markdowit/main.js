import markdownit from 'markdown-it'
import hljs from 'highlight.js' // https://highlightjs.org
import { plantuml } from "@mdit/plugin-plantuml";
import anchor from "markdown-it-anchor"
import toc from "markdown-it-table-of-contents"
// import 'highlight.js/styles/github-dark.min.css';
import HighlightJS from 'markdown-it-highlightjs'
// import 'highlight.js/styles/monokai-sublime.css'; // 引入你喜欢的主题样式文件


// Actual default values
const md = markdownit({
  highlight: function (str, lang) {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return '<pre class="hljs" style="padding:10px"><code class="language-' + lang + '">' +
          hljs.highlight(str, { language: lang, ignoreIllegals: true }).value +
          '</code></pre>';
      } catch (__) { }
    }

    return '<pre class="hljs"><code>' + md.utils.escapeHtml(str) + '</code></pre>';
  }
})

md.use(plantuml);

md.use(anchor.default); // Optional, but makes sense as you really want to link to something, see info about recommended plugins below
md.use(toc);
// const opts = {
// hljs: hljs
// }
// md.use(HighlightJS, opts)
function render(text) {
  return md.render(text);
}
function hl() {
  hljs.highlightAll()
}
var markjs = {
  render,
  hl
}
export default markjs 