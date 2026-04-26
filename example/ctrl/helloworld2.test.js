import page from "./helloworld2.js";

test("hello world", async () => {
    // given
    const render = createRender();

    // when
    page(render);

    // assert
    expect(render.size()).toBe(1);
    expect(render.get(0)).toContain("Hello world");
    expect(render.get(0)).toContain("go home");
});
