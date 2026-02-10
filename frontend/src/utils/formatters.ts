export const formatCompactNumber = (number: number): string => {
    if (!number && number !== 0) return "0";

    // For small numbers, just return with locale string
    if (number < 1000) {
        return number.toLocaleString('en-US', { maximumFractionDigits: 1 });
    }

    const suffixes = ["", "k", "M", "B", "T"];
    const suffixNum = Math.floor((("" + Math.floor(number)).length - 1) / 3);

    // Ensure we don't go out of bounds
    if (suffixNum >= suffixes.length) return number.toExponential(1);

    let shortValue = parseFloat((suffixNum !== 0 ? (number / Math.pow(1000, suffixNum)) : number).toPrecision(3));

    if (shortValue % 1 !== 0) {
        shortValue = parseFloat(shortValue.toFixed(1));
    }

    return shortValue + suffixes[suffixNum];
};
