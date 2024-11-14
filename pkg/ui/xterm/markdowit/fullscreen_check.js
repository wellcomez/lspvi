const MINIMUM_COLS = 2;
const MINIMUM_ROWS = 1;
class fullscreen_check {
    constructor(term) {
        this.term = term
        // fit.call(this)
    }
    dims = () => {
        if (!this.term) {
            return undefined;
        }

        if (!this.term.element || !this.term.element.parentElement) {
            return undefined;
        }

        // TODO: Remove reliance on private API
        const core = this.term._core;
        const dims = core._renderService.dimensions;

        if (dims.css.cell.width === 0 || dims.css.cell.height === 0) {
            return undefined;
        }
        return dims

    }
    scrollbarWidth = () => {
        const scrollbarWidth = (this.term.options.scrollback === 0
            ? 0
            : (this.term.options.overviewRuler?.width || 1));
        return scrollbarWidth
    }
    resize = (call) => {
        this.term.element.style.height = (window.innerHeight - 50) + "px"
        this.term.element.style.width = window.innerWidth + "px"
        let dims = this.check()
        // console.log("col-row", ss)
        if (dims != undefined) {
            this.term.resize(dims.cols, dims.rows);
            if (call)
                resizecall()
        }
    }
    check = () => {
        const dims = this.dims()
        const scrollbarWidth = this.scrollbarWidth()
        const parentElementStyle = window.getComputedStyle(this.term.element.parentElement);
        const parentElementHeight = parseInt(parentElementStyle.getPropertyValue('height'));
        const parentElementWidth = Math.max(0, parseInt(parentElementStyle.getPropertyValue('width')));
        const elementStyle = window.getComputedStyle(this.term.element);
        const elementPadding = {
            top: parseInt(elementStyle.getPropertyValue('padding-top')),
            bottom: parseInt(elementStyle.getPropertyValue('padding-bottom')),
            right: parseInt(elementStyle.getPropertyValue('padding-right')),
            left: parseInt(elementStyle.getPropertyValue('padding-left'))
        };
        const elementPaddingVer = elementPadding.top + elementPadding.bottom;
        const elementPaddingHor = elementPadding.right + elementPadding.left;
        const availableHeight = parentElementHeight - elementPaddingVer;
        const availableWidth = parentElementWidth - elementPaddingHor - scrollbarWidth;
        return this.fit(availableWidth, availableHeight, dims)
    }
    fit = (availableWidth, availableHeight, dims) => {
        const geometry = {
            cols: Math.max(MINIMUM_COLS, Math.floor(availableWidth / dims.css.cell.width)),
            rows: Math.max(MINIMUM_ROWS, Math.floor(availableHeight / dims.css.cell.height))
        };
        return geometry;
    }
}
export default {
    fullscreen_check
}