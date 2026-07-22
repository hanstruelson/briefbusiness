export function init() {
    document.addEventListener("DOMContentLoaded", () => {
        // scan for non-native elements
        const customElements = document.querySelectorAll(":not(:defined)");
        customElements.forEach((el) => {
            // Initialize custom elements
            console.log(`Initializing custom element: ${el.tagName}`);
        });
    });
}