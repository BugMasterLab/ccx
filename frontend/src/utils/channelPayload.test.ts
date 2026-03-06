import { describe, expect, it } from 'vitest'
import { buildChannelPayload } from './channelPayload'

describe('buildChannelPayload', () => {
  it('应序列化 reasoningMapping 与渠道级 verbosity/fastMode', () => {
    const result = buildChannelPayload({
      name: '  test-channel  ',
      serviceType: 'openai',
      baseUrl: 'https://api.example.com/v1#',
      baseUrls: [],
      website: ' https://platform.openai.com ',
      insecureSkipVerify: false,
      lowQuality: false,
      injectDummyThoughtSignature: false,
      stripThoughtSignature: false,
      description: '  desc  ',
      apiKeys: ['sk-1', '  ', 'sk-2'],
      modelMapping: { 'gpt-5': 'gpt-5.2' },
      reasoningMapping: { 'gpt-5': 'high' },
      textVerbosity: 'medium',
      fastMode: true,
      customHeaders: { 'x-test': '1' },
      proxyUrl: ' http://127.0.0.1:7890 ',
      supportedModels: ['gpt-5']
    })

    expect(result.name).toBe('test-channel')
    expect(result.baseUrl).toBe('https://api.example.com/v1')
    expect(result.website).toBe('https://platform.openai.com')
    expect(result.description).toBe('desc')
    expect(result.apiKeys).toEqual(['sk-1', 'sk-2'])
    expect(result.modelMapping).toEqual({ 'gpt-5': 'gpt-5.2' })
    expect(result.reasoningMapping).toEqual({ 'gpt-5': 'high' })
    expect(result.textVerbosity).toBe('medium')
    expect(result.fastMode).toBe(true)
    expect(result.proxyUrl).toBe('http://127.0.0.1:7890')
  })

  it('应对多个 baseUrls 去重并保留 baseUrls 输出', () => {
    const result = buildChannelPayload({
      name: 'multi',
      serviceType: 'responses',
      baseUrl: '',
      baseUrls: ['https://api.example.com/v1/', 'https://api.example.com/v1#', 'https://backup.example.com/v1'],
      website: '',
      insecureSkipVerify: false,
      lowQuality: false,
      injectDummyThoughtSignature: false,
      stripThoughtSignature: false,
      description: '',
      apiKeys: ['sk-1'],
      modelMapping: {},
      reasoningMapping: {},
      textVerbosity: '',
      fastMode: false,
      customHeaders: {},
      proxyUrl: '',
      supportedModels: []
    })

    expect(result.baseUrl).toBe('https://api.example.com/v1')
    expect(result.baseUrls).toEqual(['https://api.example.com/v1', 'https://backup.example.com/v1'])
  })
})
