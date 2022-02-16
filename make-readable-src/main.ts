import { readAll } from "https://deno.land/std@0.125.0/streams/conversion.ts";
import { createDOMPurify, JSDOM, Readability } from "./deps.ts";

function parse(url: string, html: Uint8Array) {
  const dom = new JSDOM();
  const DOMPurify = createDOMPurify(dom.window);
  const clean = DOMPurify.sanitize(new TextDecoder().decode(html));
  const doc = new JSDOM(clean, { url });
  const reader = new Readability(doc.window.document);
  const article = reader.parse();

  if (!article) {
    return { error: `No content for ${url}` };
  } else {
    return article;
  }
}

async function main(argv: string[]) {
  if (argv.length < 1) {
    return { error: "No URL argument passed" };
  }

  const url = argv[0];

  const html = await readAll(Deno.stdin);
  try {
    return parse(url, html);
  } catch (err) {
    return { error: err.toString() };
  }
}

if (import.meta.main) {
  const result = await main(Deno.args);
  console.log(JSON.stringify(result));
  Deno.exit(0);
}
