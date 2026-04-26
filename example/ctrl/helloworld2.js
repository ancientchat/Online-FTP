export default function(render) {
    render(`
        <div class="component_helloworld">
            <!-- this is all plain HTML -->
            <h1>
                Hello world
            </h1>
            <p>
                <a href="/example/" data-link>go home</a>
                <br/><br/>
                render things exactly when and how you want. After all, it's plain old javascript
             </p>
        </div>
   `);
};
