const { chromium } = require('playwright');

(async () => {
  const BASE = 'http://localhost:5174';
  const DESKTOP = { width: 1280, height: 800 };
  const MOBILE  = { width: 390,  height: 844, isMobile: true };

  const browser = await chromium.launch({ headless: true });

  async function shot(ctx, page, name, url, scroll = 0, clip = null) {
    const p = await ctx.newPage();
    await p.goto(BASE + url, { waitUntil: 'networkidle', timeout: 15000 });
    await p.waitForTimeout(600);
    if (scroll) await p.evaluate(y => window.scrollTo(0, y), scroll);
    await p.waitForTimeout(200);
    const opts = { path: `b_${name}.png` };
    if (clip) opts.clip = clip;
    else opts.fullPage = true;
    await p.screenshot(opts);
    await p.close();
    console.log(name);
  }

  // Desktop
  const dCtx = await browser.newContext({ viewport: DESKTOP });
  await shot(dCtx, null, 'd_home_top',   '/', 0, {x:0, y:0, width:1280, height:200});
  await shot(dCtx, null, 'd_home_hero',  '/', 0, {x:0, y:60, width:1280, height:500});
  await shot(dCtx, null, 'd_home_steps', '/', 700, {x:0, y:700, width:1280, height:400});
  await shot(dCtx, null, 'd_tournaments','/', 0);
  await shot(dCtx, null, 'd_login',      '/login', 0);

  // Mobile
  const mCtx = await browser.newContext({ viewport: MOBILE });
  await shot(mCtx, null, 'm_home_hero',  '/', 0, {x:0, y:0, width:390, height:400});
  await shot(mCtx, null, 'm_home_steps', '/', 500, {x:0, y:0, width:390, height:400});
  await shot(mCtx, null, 'm_home_active','/', 950, {x:0, y:0, width:390, height:300});
  await shot(mCtx, null, 'm_tournaments','/tournaments', 0);
  await shot(mCtx, null, 'm_login',      '/login', 0);

  await browser.close();
  console.log('done');
})();
