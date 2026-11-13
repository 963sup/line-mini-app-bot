# Space

Space保存可信群組身份、事件順序、獨立候選及成員讀取資格，與Repository平行。身份、Project擁有權及相容／遷移契約集中於 [獨立Project](../../../../docs/features/projects.md)；HTTP／LINE／Firestore各自維持adapter責任，不把候選當授權，不依Repository Service建立Space。
