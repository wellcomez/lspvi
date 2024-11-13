import markdownit from 'markdown-it'
import hljs from 'highlight.js' // https://highlightjs.org

// Actual default values
const md = markdownit({
    highlight: function (str, lang) {
      if (lang && hljs.getLanguage(lang)) {
        try {
          return '<pre><code class="hljs">' +
                 hljs.highlight(str, { language: lang, ignoreIllegals: true }).value +
                 '</code></pre>';
        } catch (__) {}
      }
  
      return '<pre><code class="hljs">' + md.utils.escapeHtml(str) + '</code></pre>';
    }
  });
  function render(text) {
    return md.render(text);
  }
  export default render;