# NeuroShell Real-World Experiments

This directory contains real-world experiment scripts for NeuroShell that interact with actual LLM providers using production API keys.

## Purpose

Unlike the test golden files in `test/golden/`, these scripts are designed for:
- **Real experimentation** with actual LLM providers
- **Learning and exploration** of LLM capabilities
- **Manual observation** of results and behaviors
- **Practical use case testing** with real API responses

## Directory Structure

```
examples/experiments/
├── README.md              # This file
├── gemini/               # Gemini-specific experiments
│   ├── README.md         # Gemini experiment documentation
│   └── *.neuro          # Gemini experiment scripts
└── (future providers)/   # Other provider experiments
```

## Prerequisites

1. **Build the production binary**:
   ```bash
   just build
   ```

2. **Set up API keys** in your local `.env` file:
   ```bash
   # Required for Gemini experiments
   GOOGLE_API_KEY=your_actual_gemini_api_key_here
   ```

3. **Ensure you have sufficient API credits** for the experiments

## Running Experiments

Execute experiment scripts using the batch command:

```bash
# Run a specific experiment
./bin/neuro batch examples/experiments/gemini/script-name.neuro

# Example
./bin/neuro batch examples/experiments/gemini/gemini-basic-chat.neuro
```

## Key Differences from Golden Tests

| Aspect | Golden Tests | Real Experiments |
|--------|-------------|------------------|
| **Purpose** | Automated regression testing | Manual exploration and learning |
| **API Keys** | Mocked/fake responses | Real API keys from .env |
| **Execution** | Test framework automation | Manual batch command execution |
| **Output** | Expected vs actual comparison | Manual observation and analysis |
| **Focus** | Ensuring consistent behavior | Discovering capabilities and patterns |

## Best Practices

1. **Monitor API Usage**: These scripts make real API calls that consume credits
2. **Start Small**: Begin with basic experiments before complex workflows
3. **Document Learnings**: Keep notes on interesting behaviors or findings
4. **Clean Up**: Scripts should clean up created models/sessions when possible
5. **Error Handling**: Include proper error handling for real-world scenarios

## Safety Notes

- **Never commit API keys** to version control
- **Be mindful of rate limits** when running experiments
- **Monitor costs** associated with API usage
- **Test incrementally** to avoid unexpected high usage

## Contributing

When adding new experiment scripts:
1. Follow the naming convention: `provider-experiment-type.neuro`
2. Include comments explaining the experiment's purpose
3. Add cleanup steps where appropriate
4. Document expected outcomes and key observations