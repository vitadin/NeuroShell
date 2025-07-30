# OpenAI Experiments

This directory contains real-world experiment scripts specifically for OpenAI's LLM models.

## Available Experiments

### Basic Experiments
- **`openai-basic-chat.neuro`** - Simple model creation and basic chat interaction with GPT-4.1

## Prerequisites

1. **OpenAI API Key**: Set `OPENAI_API_KEY` in your `.env` file
2. **Production Build**: Run `just build` to create `./bin/neuro`
3. **API Credits**: Ensure you have sufficient OpenAI API credits

## OpenAI Model Catalog IDs

Available OpenAI models in the system:
- **G41** - GPT-4.1 (Balanced performance and capability)
- **G4O** - GPT-4o (Versatile, high-intelligence flagship)
- **O3** - O3 (Advanced reasoning and multi-step problem solving)
- **O4M** - O4 Mini (Efficient, cost-effective)  
- **O1** - O1 (Reasoning-focused model)
- **O1P** - O1 Pro (Premium reasoning capabilities)
- **O3P** - O3 Pro (Premium reasoning with extended capabilities)

## Key OpenAI Features to Experiment With

### Reasoning Capabilities (O-series models)
- **Reasoning Effort**: Control reasoning complexity (low, medium, high)
  - `low`: Quick, efficient responses
  - `medium`: Balanced reasoning and speed (default)
  - `high`: Deep reasoning for complex problems

### Parameters
- **Temperature**: Creativity/randomness (0.0-2.0)
- **Max Tokens**: Response length limit
- **Max Output Tokens**: Total output including reasoning tokens (for reasoning models)
- **Reasoning Summary**: Enable reasoning summaries (auto, detailed, concise)

## Example Usage

```bash
# Run basic chat experiment with GPT-4.1
./bin/neuro batch examples/experiments/openai/openai-basic-chat.neuro
```

## Experiment Design Patterns

### 1. Model Creation Pattern
```neuro
%% Create OpenAI model with specific configuration
\openai-model-new[catalog_id="G41", temperature="0.7"] my-experiment-model
\model-activate my-experiment-model
```

### 2. Reasoning Model Pattern (for O-series)
```neuro
%% Create reasoning model with effort control
\openai-model-new[catalog_id="O3", reasoning_effort="medium", max_output_tokens="10000"] reasoning-model
\model-activate reasoning-model
```

### 3. Cleanup Pattern
```neuro
%% Always clean up created models
\model-delete my-experiment-model
```

## Observations to Track

When running experiments, observe:
1. **Response Quality**: How different models handle various tasks
2. **Performance**: Response time differences between models
3. **Reasoning**: How O-series models show their reasoning process
4. **Consistency**: Reproducibility of responses with same parameters
5. **Error Patterns**: Common failure modes and how to handle them

## Cost Considerations

- **GPT-4.1 (G41)**: Balanced cost and performance for general use
- **GPT-4o (G4O)**: Higher cost but versatile capabilities
- **O-series models**: Variable cost based on reasoning effort and output tokens
- **Reasoning effort**: Higher effort levels increase API costs
- **Long conversations**: Consider session management for cost control

## Troubleshooting

### Common Issues
1. **API Key Not Found**: Ensure `OPENAI_API_KEY` is set in `.env`
2. **Rate Limiting**: Space out requests if hitting rate limits
3. **Model Not Found**: Check catalog IDs are correct (G41, G4O, O3, etc.)
4. **Quota Exceeded**: Monitor your OpenAI API usage and quotas
5. **Token Limits**: Be aware of context window limits for different models

### Debug Commands
```bash
# Check current model status
./bin/neuro batch -c "\model-status"

# Verify API configuration
./bin/neuro batch -c "\llm-api-show[provider=openai]"

# List available OpenAI models
./bin/neuro batch -c "\model-catalog[provider=openai]"
```

## Model Selection Guide

- **G41 (GPT-4.1)**: Best for general-purpose tasks, balanced performance
- **G4O (GPT-4o)**: Best for complex tasks requiring high intelligence
- **O3**: Best for multi-step reasoning and analysis
- **O1/O1P**: Best for mathematical and logical reasoning
- **O4M**: Best for cost-effective, simple tasks

Choose based on your specific use case, budget, and performance requirements.