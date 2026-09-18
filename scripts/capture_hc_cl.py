#!/usr/bin/env python3
"""
Capturer HC et CL depuis script Akamai exécuté dans un vrai navigateur.
Utilise camoufox pour éviter les blocages Akamai.

Usage:
    python3 capture_hc_cl.py --url "https://www.zalando.fr/..." --output hc_cl.json
"""

import sys
import json
import asyncio
import re
from pathlib import Path

try:
    from playwright.async_api import async_playwright
except ImportError:
    print("Erreur: pip install playwright")
    sys.exit(1)


async def capturer_hc_cl(url: str) -> dict:
    """Capture HC et CL depuis script Akamai via Playwright."""
    
    async with async_playwright() as p:
        # Utiliser Chromium avec camoufox ou autre
        browser = await p.chromium.launch(headless=True)
        context = await browser.new_context()
        page = await context.new_page()
        
        # Logger toutes les exécutions JS globales
        hc_value = None
        cl_value = None
        
        async def capture_execution(source: str) -> None:
            """Intercepter les assignations de variables globales."""
            nonlocal hc_value, cl_value
            
            # Chercher HC = ou var HC =
            if "HC=" in source or "var HC" in source:
                match = re.search(r'HC\s*=\s*(\d+)', source)
                if match:
                    hc_value = int(match.group(1))
                    print(f"[CAPTURE] HC trouvé: {hc_value}")
            
            # Chercher CL = [...]
            if "CL=" in source or "var CL" in source:
                match = re.search(r'CL\s*=\s*\[(.*?)\]', source)
                if match:
                    try:
                        cl_str = "[" + match.group(1) + "]"
                        cl_value = json.loads(cl_str)
                        print(f"[CAPTURE] CL trouvé: {cl_value}")
                    except:
                        pass
        
        # Charger la page
        print(f"Chargement: {url}")
        await page.goto(url, wait_until="networkidle")
        
        # Attendre le script Akamai
        await page.wait_for_function(
            "typeof HC !== 'undefined' && typeof CL !== 'undefined'",
            timeout=10000
        )
        
        # Extraire HC et CL depuis le scope global
        hc_value = await page.evaluate("typeof HC !== 'undefined' ? HC : null")
        cl_value = await page.evaluate("typeof CL !== 'undefined' ? CL : null")
        
        print(f"\n✓ Capturation complète:")
        print(f"  HC: {hc_value}")
        print(f"  CL: {cl_value}")
        
        await browser.close()
        
        return {
            "hc": hc_value,
            "cl": cl_value,
            "url": url
        }


async def main():
    import argparse
    
    parser = argparse.ArgumentParser(description="Capturer HC/CL depuis Akamai")
    parser.add_argument("--url", required=True, help="URL Zalando")
    parser.add_argument("--output", default="hc_cl.json", help="Fichier de sortie")
    
    args = parser.parse_args()
    
    result = await capturer_hc_cl(args.url)
    
    with open(args.output, "w") as f:
        json.dump(result, f, indent=2)
    
    print(f"\n✓ Sauvegardé dans: {args.output}")


if __name__ == "__main__":
    if sys.platform == "win32":
        asyncio.set_event_loop_policy(asyncio.WindowsSelectorEventLoopPolicy())
    
    asyncio.run(main())
