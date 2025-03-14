const FacetCutAction = { Add: 0, Replace: 1, Remove: 2 };

function getSelectors(contract) {
    const signatures = Object.keys(contract.interface.functions);
    const selectors = signatures.reduce((acc, val) => {
        if (val !== 'init(bytes)') {
            acc.push(contract.interface.getSighash(val));
        }
        return acc;
    }, []);
    return selectors;
}

module.exports = {
    FacetCutAction,
    getSelectors
}; 