import markdownit from 'markdown-it'
import hljs from 'highlight.js' // https://highlightjs.org
import { plantuml } from "@mdit/plugin-plantuml";
import axios from 'axios'
import anchor from "markdown-it-anchor"
import toc from "markdown-it-table-of-contents"
// import 'highlight.js/styles/github-dark.min.css';
// import HighlightJS from 'markdown-it-highlightjs'
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
// md.renderer.rules.paragraph_open = function (tokens, idx, options, env, self) {
//   return '<p class="markdown-class">';
// };

// // 自定义段落关闭标签
// md.renderer.rules.paragraph_close = function (tokens, idx, options, env, self) {
//   return '</p>';
// };

// // 自定义标题渲染器
// md.renderer.rules.heading_open = function (tokens, idx, options, env, self) {
//   const token = tokens[idx];
//   const level = token.tag.slice(1);
//   return `<h${level} class="markdown-heading">`;
// };

// // 自定义标题关闭标签
// md.renderer.rules.heading_close = function (tokens, idx, options, env, self) {
//   const token = tokens[idx];
//   const level = token.tag.slice(1);
//   return `</h${level}>`;
// };
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
// function hl() {
//   hljs.highlightAll()
// }
function render_file(mdfile) {
  let url = "https://" + window.location.host + "/" + mdfile
  axios.get(url, { responseType: "text" }).then((resp) => {
    let ss = render(resp.data)
    let title = mdfile.split("/").pop()
    document.body.innerHTML = ss
    document.title = title
    // new Term()
    // markjs.hl()
  });
}
export default {
  render_file
}