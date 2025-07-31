# Gemini Experiments

This directory contains real-world experiment scripts specifically for Google's Gemini LLM models.

## Available Experiments

### Basic Experiments
- **`gemini-basic-chat.neuro`** - Simple model creation and basic chat interaction
- **`gemini-thinking-experiment.neuro`** - Explore Gemini's thinking capabilities with different budgets
- **`gemini-parameter-tuning.neuro`** - Test various model parameters (temperature, max_tokens, etc.)

### Advanced Experiments  
- **`gemini-model-comparison.neuro`** - Compare different Gemini models (Flash, Pro, Lite)
- **`gemini-workflow-automation.neuro`** - End-to-end workflow automation with model management

## Prerequisites

1. **Google API Key**: Set `GOOGLE_API_KEY` in your `.env` file
2. **Production Build**: Run `just build` to create `./bin/neuro`
3. **API Credits**: Ensure you have sufficient Gemini API credits

## Gemini Model Catalog IDs

Available Gemini models in the system:
- **GM25F** - Gemini 2.5 Flash (Fast, cost-effective)
- **GM25P** - Gemini 2.5 Pro (High-performance, advanced reasoning)
- **GM25FL** - Gemini 2.5 Flash Lite (Lightweight, basic tasks)

## Key Gemini Features to Experiment With

### Thinking Capabilities
- **Thinking Budget**: Control how much the model "thinks" before responding
  - `-1`: Unlimited thinking (model decides)
  - `0`: No thinking (direct response)
  - `1024-16384`: Specific thinking token limits

### Parameters
- **Temperature**: Creativity/randomness (0.0-2.0)
- **Max Tokens**: Response length limit
- **Top P**: Nucleus sampling parameter
- **Top K**: Top-k sampling parameter

## Example Usage

```bash
# Run basic chat experiment
./bin/neuro batch examples/experiments/gemini/gemini-basic-chat.neuro

# Run thinking experiment  
./bin/neuro batch examples/experiments/gemini/gemini-thinking-experiment.neuro

# Compare different models
./bin/neuro batch examples/experiments/gemini/gemini-model-comparison.neuro
```

## Experiment Design Patterns

### 1. Model Creation Pattern
```neuro
%% Create Gemini model with specific configuration
\gemini-model-new[catalog_id="GM25F", thinking_budget="2048", temperature="0.7"] my-experiment-model
\model-activate my-experiment-model
```

### 2. Parameter Testing Pattern
```neuro
%% Test different temperature settings
\set[temperatures="0.1,0.5,0.9,1.5"]
%% Iterate through parameters and observe differences
```

### 3. Cleanup Pattern
```neuro
%% Always clean up created models
\model-delete my-experiment-model
```

## Observations to Track

When running experiments, observe:
1. **Response Quality**: How does thinking budget affect response depth?
2. **Performance**: Response time differences between models
3. **Creativity**: How temperature affects creative vs factual responses
4. **Consistency**: Reproducibility of responses with same parameters
5. **Error Patterns**: Common failure modes and how to handle them

## Cost Considerations

- **Flash models**: Most cost-effective for basic tasks
- **Pro models**: Higher cost but better reasoning capabilities
- **Thinking budget**: Higher budgets increase API costs
- **Long conversations**: Consider session management for cost control

## Troubleshooting

### Common Issues
1. **API Key Not Found**: Ensure `GOOGLE_API_KEY` is set in `.env`
2. **Rate Limiting**: Space out requests if hitting rate limits
3. **Model Not Found**: Check catalog IDs are correct (GM25F, GM25P, GM25FL)
4. **Quota Exceeded**: Monitor your Gemini API usage and quotas

### Debug Commands
```bash
# Check current model status
./bin/neuro batch -c "\model-status"

# Verify API configuration
./bin/neuro batch -c "\llm-api-load[provider=gemini]"
```