<script setup>
// Step 1 — Welcome (spec §5.4). Pure hero content: the single CTA is the
// wizard footer's Continue ("Get started") — no hero button (F21).
</script>

<template>
    <div>
        <div class="pc-panel-enter">
            <h1 class="setup-title">{{ $t('welcome.title') }}</h1>
            <p class="setup-lede">{{ $t('welcome.lede') }}</p>
        </div>

        <div class="setup-panels">
            <!-- Mini signal-path diagram (brand pigments at icon size only) -->
            <div class="pc-panel welcome-diagram pc-panel-enter pc-panel-enter--2" :aria-label="$t('welcome.diagramAria')">
                <span class="diagram-node">
                    <svg class="diagram-glyph diagram-glyph--plex" width="24" height="24" viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
                        <polyline
                            fill="none"
                            stroke="currentColor"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2.5"
                            points="4.5 24 23.444 24 12.808 9.342 16.883 9.342 27.519 24 16.883 38.658 20.957 38.658 31.594 24 20.957 9.342 25.032 9.342 35.668 24 25.032 38.658 29.107 38.658 39.743 24 43.5 24"
                        />
                    </svg>
                    <span class="diagram-label">{{ $t('welcome.plex') }}</span>
                </span>
                <span class="diagram-line" aria-hidden="true"></span>
                <span class="diagram-node">
                    <span class="diagram-name">PlexCord</span>
                </span>
                <span class="diagram-line" aria-hidden="true"></span>
                <span class="diagram-node">
                    <svg class="diagram-glyph diagram-glyph--discord" width="24" height="24" viewBox="0 0 71 55" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
                        <path
                            d="M60.1045 4.8978C55.5792 2.8214 50.7265 1.2916 45.6527 0.41542C45.5603 0.39851 45.468 0.440769 45.4204 0.525289C44.7963 1.6353 44.105 3.0834 43.6209 4.2216C38.1637 3.4046 32.7345 3.4046 27.3892 4.2216C26.905 3.0581 26.1886 1.6353 25.5617 0.525289C25.5141 0.443589 25.4218 0.40133 25.3294 0.41542C20.2584 1.2888 15.4057 2.8186 10.8776 4.8978C10.8384 4.9147 10.8048 4.9429 10.7825 4.9795C1.57795 18.7309 -0.943561 32.1443 0.293408 45.3914C0.299005 45.4562 0.335386 45.5182 0.385761 45.5576C6.45866 50.0174 12.3413 52.7249 18.1147 54.5195C18.2071 54.5477 18.305 54.5139 18.3638 54.4378C19.7295 52.5728 20.9469 50.6063 21.9907 48.5383C22.0523 48.4172 21.9935 48.2735 21.8676 48.2256C19.9366 47.4931 18.0979 46.6 16.3292 45.5858C16.1893 45.5041 16.1781 45.304 16.3068 45.2082C16.679 44.9293 17.0513 44.6391 17.4067 44.3461C17.471 44.2926 17.5606 44.2813 17.6362 44.3151C29.2558 49.6202 41.8354 49.6202 53.3179 44.3151C53.3935 44.2785 53.4831 44.2898 53.5502 44.3433C53.9057 44.6363 54.2779 44.9293 54.6529 45.2082C54.7816 45.304 54.7732 45.5041 54.6333 45.5858C52.8646 46.6197 51.0259 47.4931 49.0921 48.2228C48.9662 48.2707 48.9102 48.4172 48.9718 48.5383C50.038 50.6034 51.2554 52.5699 52.5959 54.435C52.6519 54.5139 52.7526 54.5477 52.845 54.5195C58.6464 52.7249 64.529 50.0174 70.6019 45.5576C70.6551 45.5182 70.6887 45.459 70.6943 45.3942C72.1747 30.0791 68.2147 16.7757 60.1968 4.9823C60.1772 4.9429 60.1437 4.9147 60.1045 4.8978ZM23.7259 37.3253C20.2276 37.3253 17.3451 34.1136 17.3451 30.1693C17.3451 26.225 20.1717 23.0133 23.7259 23.0133C27.308 23.0133 30.1626 26.2532 30.1066 30.1693C30.1066 34.1136 27.28 37.3253 23.7259 37.3253ZM47.3178 37.3253C43.8196 37.3253 40.9371 34.1136 40.9371 30.1693C40.9371 26.225 43.7636 23.0133 47.3178 23.0133C50.9 23.0133 53.7545 26.2532 53.6986 30.1693C53.6986 34.1136 50.9 37.3253 47.3178 37.3253Z"
                            fill="currentColor"
                        />
                    </svg>
                    <span class="diagram-label">{{ $t('welcome.discord') }}</span>
                </span>
            </div>

            <!-- What you'll need: two 44px rows with muted check glyphs -->
            <div class="pc-panel welcome-checklist pc-panel-enter pc-panel-enter--3">
                <span class="pc-eyebrow">{{ $t('welcome.whatYouNeed') }}</span>
                <ul class="checklist">
                    <li class="checklist-row">
                        <i class="pi pi-check checklist-glyph" aria-hidden="true"></i>
                        <span class="checklist-texts">
                            <span class="checklist-label">{{ $t('welcome.needPlexTitle') }}</span>
                            <span class="checklist-caption">{{ $t('welcome.needPlexCaption') }}</span>
                        </span>
                    </li>
                    <li class="checklist-row">
                        <i class="pi pi-check checklist-glyph" aria-hidden="true"></i>
                        <span class="checklist-texts">
                            <span class="checklist-label">{{ $t('welcome.needDiscordTitle') }}</span>
                            <span class="checklist-caption">{{ $t('welcome.needDiscordCaption') }}</span>
                        </span>
                    </li>
                </ul>
            </div>
        </div>
    </div>
</template>

<style scoped>
/* ---- Mini signal-path diagram ---- */
.welcome-diagram {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 12px;
}
.diagram-node {
    display: inline-flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
}
.diagram-glyph--plex {
    color: var(--pc-plex);
}
.diagram-glyph--discord {
    color: var(--pc-blurple);
}
.diagram-label {
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
}
.diagram-name {
    font-size: 13px;
    font-weight: 600;
    color: var(--pc-text);
}
.diagram-line {
    flex: 1;
    max-width: 96px;
    height: 1px;
    background: var(--pc-border);
}

/* ---- Checklist rows (44px, hairline-divided) ---- */
.welcome-checklist .pc-eyebrow {
    display: block;
    margin-bottom: 12px;
}
.checklist {
    list-style: none;
    margin: 0;
    padding: 0;
}
.checklist-row {
    display: flex;
    align-items: center;
    gap: 12px;
    min-height: 44px;
    padding: 12px 0;
}
.checklist-row + .checklist-row {
    border-top: 1px solid var(--pc-border-subtle);
}
.checklist-glyph {
    font-size: 14px;
    color: var(--pc-text-muted);
    flex: none;
}
.checklist-texts {
    display: flex;
    flex-direction: column;
    gap: 2px;
}
.checklist-label {
    font-size: var(--pc-text-body);
    color: var(--pc-text);
}
.checklist-caption {
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
}
</style>
