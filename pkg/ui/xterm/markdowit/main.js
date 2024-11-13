import markdownit from 'markdown-it'
import hljs from 'highlight.js' // https://highlightjs.org
import { plantuml } from "@mdit/plugin-plantuml";

// Actual default values
const md = markdownit({
  highlight: function (str, lang) {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return '<pre><code class="hljs">' +
          hljs.highlight(str, { language: lang, ignoreIllegals: true }).value +
          '</code></pre>';
      } catch (__) { }
    }

    return '<pre><code class="hljs">' + md.utils.escapeHtml(str) + '</code></pre>';
  }
});
md.use(plantuml);
function render(text) {
  return md.render(text);
}
function hl() {
  hljs.highlightAll()
}
var markjs= {
  render,
  hl
}
export default markjs 