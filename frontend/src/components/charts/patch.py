import re

with open('c:\\Pham_Phuc\\UTE\\YEAR4\\HK2\\Do_an_tot_nghiep_2425\\Proj\\Code\\code\\Shundao\\Shundao\\Shundao\\frontend\\src\\components\\charts\\ProductionChart.tsx', 'r', encoding='utf-8') as f:
    text = f.read()

target1 = r'''    // Memoize: Total MWh per site\n    const \{ totalS1MWh, totalS2MWh \} = useMemo\(\(\) => \{\n        if \(viewMode === 'day'\) \{\n            const validPoints = data\.filter\(d =>\n                \(d\.site1DailyEnergy != null && d\.site1DailyEnergy > 0\) \|\|\n                \(d\.site2DailyEnergy != null && d\.site2DailyEnergy > 0\)\n            \);\n            return \{\n                totalS1MWh: \(validPoints\.reduce\(\(sum, d\) => sum \+ \(d\.site1DailyEnergy \|\| 0\), 0\) \* \(5 / 60\)\) / 1000,\n                totalS2MWh: \(validPoints\.reduce\(\(sum, d\) => sum \+ \(d\.site2DailyEnergy \|\| 0\), 0\) \* \(5 / 60\)\) / 1000,\n            \};\n        \}\n        const monthly = monthlyData \|\| \[\];\n        return \{\n            totalS1MWh: monthly\.reduce\(\(sum, d\) => sum \+ \(d\.site1MaxPower \|\| 0\), 0\) / 1000,\n            totalS2MWh: monthly\.reduce\(\(sum, d\) => sum \+ \(d\.site2MaxPower \|\| 0\), 0\) / 1000,\n        \};\n    \}, \[data, viewMode, monthlyData\]\);'''

replace1 = '''    // Memoize: Current/Peak Power per site
    const { powerS1, powerS2, unitLabel, labelPrefixS1, labelPrefixS2 } = useMemo(() => {
        if (viewMode === 'day') {
            let s1 = 0;
            let s2 = 0;
            for (let i = data.length - 1; i >= 0; i--) {
                if (data[i].site1DailyEnergy != null && data[i].site1DailyEnergy! > 0 && s1 === 0) {
                    s1 = data[i].site1DailyEnergy! / 1000;
                }
                if (data[i].site2DailyEnergy != null && data[i].site2DailyEnergy! > 0 && s2 === 0) {
                    s2 = data[i].site2DailyEnergy! / 1000;
                }
                if (s1 > 0 && s2 > 0) break;
            }
            return {
                powerS1: s1,
                powerS2: s2,
                unitLabel: "MW",
                labelPrefixS1: "Công su?t th?i gian th?c Shundao1:",
                labelPrefixS2: "Công su?t th?i gian th?c Shundao2:"
            };
        }
        const monthly = monthlyData || [];
        return {
            powerS1: monthly.length > 0 ? Math.max(...monthly.map(d => d.site1MaxPower || 0)) / 1000 : 0,
            powerS2: monthly.length > 0 ? Math.max(...monthly.map(d => d.site2MaxPower || 0)) / 1000 : 0,
            unitLabel: "MW",
            labelPrefixS1: "Công su?t d?nh Shundao1:",
            labelPrefixS2: "Công su?t d?nh Shundao2:"
        };
    }, [data, viewMode, monthlyData]);'''

target2 = r'''                                    \{visibleSites\.site1 && \(\n                                        <div className="flex items-baseline gap-1">\n                                            <span className="text-\[14px\] font-bold text-blue-400">T?ng công su?t Shundao1:</span>\n                                            <span className="text-\[16px\] font-semibold text-black">\{totalS1MWh\.toFixed\(2\)\} MWh</span>\n                                        </div>\n                                    \)\}\n                                    \{visibleSites\.site2 && \(\n                                        <div className="flex items-baseline gap-1">\n                                            <span className="text-\[14px\] font-bold text-blue-400">T?ng công su?t Shundao2:</span>\n                                            <span className="text-\[16px\] font-semibold text-black">\{totalS2MWh\.toFixed\(2\)\} MWh</span>\n                                        </div>\n                                    \)\}'''

replace2 = '''                                    {visibleSites.site1 && (
                                        <div className="flex items-baseline gap-1">
                                            <span className="text-[14px] font-bold text-blue-400">{labelPrefixS1}</span>
                                            <span className="text-[16px] font-semibold text-black">{powerS1.toFixed(3)} {unitLabel}</span>
                                        </div>
                                    )}
                                    {visibleSites.site2 && (
                                        <div className="flex items-baseline gap-1">
                                            <span className="text-[14px] font-bold text-blue-400">{labelPrefixS2}</span>
                                            <span className="text-[16px] font-semibold text-black">{powerS2.toFixed(3)} {unitLabel}</span>
                                        </div>
                                    )}'''

text = re.sub(target1.replace('\\n', r'\\r?\\n'), replace1, text, flags=re.MULTILINE)
text = re.sub(target2.replace('\\n', r'\\r?\\n'), replace2, text, flags=re.MULTILINE)

with open('c:\\Pham_Phuc\\UTE\\YEAR4\\HK2\\Do_an_tot_nghiep_2425\\Proj\\Code\\code\\Shundao\\Shundao\\Shundao\\frontend\\src\\components\\charts\\ProductionChart.tsx', 'w', encoding='utf-8') as f:
    f.write(text)

print("Python replace done")
